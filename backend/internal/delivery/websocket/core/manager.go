package core

import (
	"sync"
	"terraforming-mars-backend/internal/logger"
	"unsafe"

	"go.uber.org/zap"
)

// Manager handles WebSocket connection lifecycle and organization
type Manager struct {
	connections     map[*Connection]bool
	gameConnections map[string]map[*Connection]bool
	mu              sync.RWMutex
	logger          *zap.Logger
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		connections:     make(map[*Connection]bool),
		gameConnections: make(map[string]map[*Connection]bool),
		logger:          logger.Get(),
	}
}

// RegisterConnection registers a new connection
func (m *Manager) RegisterConnection(connection *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[connection] = true
	m.logger.Debug("🔗 Client connected to server", zap.String("connection_id", connection.ID))
}

// UnregisterConnection unregisters a connection and handles cleanup.
// Returns connection metadata including whether it was a spectator.
func (m *Manager) UnregisterConnection(connection *Connection) (playerID, spectatorID, gameID string, connType ConnectionType, shouldBroadcast bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[connection]; !exists {
		return "", "", "", ConnectionTypePlayer, false
	}

	delete(m.connections, connection)
	connection.CloseSend()

	playerID, gameID = connection.GetPlayer()
	spectatorID = connection.SpectatorID
	connType = connection.ConnType

	if connType == ConnectionTypeSpectator {
		shouldBroadcast = gameID != "" && spectatorID != ""
	} else {
		shouldBroadcast = gameID != "" && playerID != ""
	}

	// Remove from game connections
	if gameConns, exists := m.gameConnections[gameID]; exists {
		delete(gameConns, connection)

		if len(gameConns) == 0 {
			delete(m.gameConnections, gameID)
			m.logger.Debug("Removed empty game connections map", zap.String("game_id", gameID))
		}
	}

	connection.Close()

	m.logger.Debug("⛓️‍💥 Client disconnected from server",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("spectator_id", spectatorID),
		zap.String("game_id", gameID),
		zap.String("conn_type", string(connType)))

	return playerID, spectatorID, gameID, connType, shouldBroadcast
}

// AddToGame adds a connection to a game group
func (m *Manager) AddToGame(connection *Connection, gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.gameConnections[gameID] == nil {
		m.gameConnections[gameID] = make(map[*Connection]bool)
	}
	m.gameConnections[gameID][connection] = true
}

// GetGameConnections returns all connections for a specific game (read-only copy)
func (m *Manager) GetGameConnections(gameID string) map[*Connection]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameConns := m.gameConnections[gameID]
	if gameConns == nil {
		return nil
	}

	connections := make(map[*Connection]bool, len(gameConns))
	for conn := range gameConns {
		connections[conn] = true
	}
	return connections
}

// GetConnectionCount returns the total number of registered connections
func (m *Manager) GetConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// RemoveExistingPlayerConnection removes any existing connection for the given player
// This is used during reconnection to clean up old connections before adding new ones
// CRITICAL: excludeConnection should be the current connection making the request to avoid cleaning it up
func (m *Manager) RemoveExistingPlayerConnection(playerID, gameID string, excludeConnection *Connection) *Connection {
	m.mu.Lock()
	defer m.mu.Unlock()

	var existingConnection *Connection
	var matchingConnections []*Connection

	m.logger.Debug("🔍 Starting connection cleanup search",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("exclude_connection_id", excludeConnection.ID),
		zap.Uintptr("exclude_connection_ptr", uintptr(unsafe.Pointer(excludeConnection))))

	for connection := range m.connections {
		existingPlayerID, existingGameID := connection.GetPlayer()
		if existingPlayerID == playerID && existingGameID == gameID {
			matchingConnections = append(matchingConnections, connection)

			m.logger.Debug("🔎 Found matching connection",
				zap.String("connection_id", connection.ID),
				zap.Uintptr("connection_ptr", uintptr(unsafe.Pointer(connection))),
				zap.Bool("is_excluded", connection == excludeConnection),
				zap.String("player_id", existingPlayerID),
				zap.String("game_id", existingGameID))

			if connection != excludeConnection {
				existingConnection = connection
				break
			}
		}
	}

	m.logger.Debug("🔎 Connection search complete",
		zap.Int("total_matching", len(matchingConnections)),
		zap.Bool("found_to_cleanup", existingConnection != nil))

	if existingConnection == nil {
		m.logger.Debug("🔍 No existing connection to clean up for reconnecting player",
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.String("current_connection_id", excludeConnection.ID))
		return nil
	}

	m.logger.Debug("🧹 Cleaning up existing connection for reconnecting player",
		zap.String("existing_connection_id", existingConnection.ID),
		zap.String("current_connection_id", excludeConnection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Uintptr("existing_connection_ptr", uintptr(unsafe.Pointer(existingConnection))),
		zap.Uintptr("current_connection_ptr", uintptr(unsafe.Pointer(excludeConnection))))

	delete(m.connections, existingConnection)
	existingConnection.CloseSend()

	// Remove from game connections
	if gameConns, exists := m.gameConnections[gameID]; exists {
		delete(gameConns, existingConnection)

		if len(gameConns) == 0 {
			delete(m.gameConnections, gameID)
			m.logger.Debug("Removed empty game connections map after cleanup", zap.String("game_id", gameID))
		}
	}

	existingConnection.Close()

	m.logger.Debug("✅ Existing connection cleaned up for reconnecting player",
		zap.String("old_connection_id", existingConnection.ID),
		zap.String("current_connection_id", excludeConnection.ID),
		zap.String("player_id", playerID))

	return existingConnection
}

// CloseAllConnections closes all active connections
func (m *Manager) CloseAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("🛑 Closing all active connections", zap.Int("connection_count", len(m.connections)))

	for connection := range m.connections {
		connection.Close()
	}

	m.connections = make(map[*Connection]bool)
	m.gameConnections = make(map[string]map[*Connection]bool)

	m.logger.Info("⛓️‍💥 All client connections closed by server")
}

// GetConnectionBySpectatorID finds a connection for a specific spectator in a game.
func (m *Manager) GetConnectionBySpectatorID(gameID, spectatorID string) *Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameConnections, exists := m.gameConnections[gameID]
	if !exists {
		return nil
	}

	for connection := range gameConnections {
		if connection.SpectatorID == spectatorID && connection.IsSpectator() {
			return connection
		}
	}

	return nil
}

// GetSpectatorConnections returns all spectator connections for a game.
func (m *Manager) GetSpectatorConnections(gameID string) []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var spectators []*Connection
	for conn := range m.gameConnections[gameID] {
		if conn.IsSpectator() {
			spectators = append(spectators, conn)
		}
	}
	return spectators
}

// GetConnectionByPlayerID finds a connection for a specific player in a game
func (m *Manager) GetConnectionByPlayerID(gameID, playerID string) *Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameConnections, exists := m.gameConnections[gameID]
	if !exists {
		return nil
	}

	for connection := range gameConnections {
		if connection.PlayerID == playerID {
			return connection
		}
	}

	return nil
}
