package core

import (
	"sync"
)

// Manager indexes open connections by game. It is safe for concurrent use: broadcasts
// can come from bot goroutines while the hub binds and removes connections.
type Manager struct {
	mu              sync.RWMutex
	connections     map[*Connection]struct{}
	gameConnections map[string]map[*Connection]struct{}
}

// NewManager creates an empty connection index.
func NewManager() *Manager {
	return &Manager{
		connections:     make(map[*Connection]struct{}),
		gameConnections: make(map[string]map[*Connection]struct{}),
	}
}

func (m *Manager) add(c *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[c] = struct{}{}
}

// remove drops a connection and returns the identity it had, or false when it was
// already removed.
func (m *Manager) remove(c *Connection) (Identity, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connections[c]; !ok {
		return Identity{}, false
	}
	delete(m.connections, c)
	previous := c.setIdentity(Identity{})
	m.leaveGameLocked(c, previous.GameID)
	return previous, true
}

// bind gives a registered connection a new identity and returns the one it had.
func (m *Manager) bind(c *Connection, id Identity) (Identity, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connections[c]; !ok {
		return Identity{}, false
	}
	previous := c.setIdentity(id)
	m.leaveGameLocked(c, previous.GameID)
	if id.GameID != "" {
		if m.gameConnections[id.GameID] == nil {
			m.gameConnections[id.GameID] = make(map[*Connection]struct{})
		}
		m.gameConnections[id.GameID][c] = struct{}{}
	}
	return previous, true
}

func (m *Manager) leaveGameLocked(c *Connection, gameID string) {
	conns, ok := m.gameConnections[gameID]
	if !ok {
		return
	}
	delete(conns, c)
	if len(conns) == 0 {
		delete(m.gameConnections, gameID)
	}
}

// GameConnections returns every connection in a game.
func (m *Manager) GameConnections(gameID string) []*Connection {
	return m.matching(gameID, func(Identity) bool { return true })
}

// PlayerConnections returns every connection acting as the player, one per device.
func (m *Manager) PlayerConnections(gameID, playerID string) []*Connection {
	return m.matching(gameID, func(id Identity) bool { return id.PlayerID == playerID })
}

// SpectatorConnections returns every spectator connection in a game.
func (m *Manager) SpectatorConnections(gameID string) []*Connection {
	return m.matching(gameID, func(id Identity) bool { return id.SpectatorID != "" })
}

// SpectatorConnection returns the connection of one spectator, or nil.
func (m *Manager) SpectatorConnection(gameID, spectatorID string) *Connection {
	conns := m.matching(gameID, func(id Identity) bool { return id.SpectatorID == spectatorID })
	if len(conns) == 0 {
		return nil
	}
	return conns[0]
}

func (m *Manager) matching(gameID string, match func(Identity) bool) []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Connection
	for c := range m.gameConnections[gameID] {
		if match(c.Identity()) {
			result = append(result, c)
		}
	}
	return result
}

// ConnectionCount returns the number of open connections.
func (m *Manager) ConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

func (m *Manager) closeAll() {
	m.mu.Lock()
	conns := make([]*Connection, 0, len(m.connections))
	for c := range m.connections {
		conns = append(conns, c)
	}
	m.connections = make(map[*Connection]struct{})
	m.gameConnections = make(map[string]map[*Connection]struct{})
	m.mu.Unlock()
	for _, c := range conns {
		c.Close()
	}
}
