package core

import (
	"sync"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (64KB for game state updates)
	maxMessageSize = 64 * 1024
)

// ConnectionType distinguishes player connections from spectator connections.
type ConnectionType string

const (
	ConnectionTypePlayer    ConnectionType = "player"
	ConnectionTypeSpectator ConnectionType = "spectator"
)

// Connection represents a WebSocket connection
type Connection struct {
	ID          string
	PlayerID    string
	SpectatorID string
	GameID      string
	ConnType    ConnectionType
	Conn        *websocket.Conn
	Send        chan dto.WebSocketMessage

	// Callbacks for hub communication
	onMessage    func(HubMessage)
	onDisconnect func(*Connection)

	// Direct reference to manager for game association
	manager *Manager

	// Synchronization
	mu         sync.RWMutex
	logger     *zap.Logger
	Done       chan struct{}
	closeOnce  sync.Once
	sendClosed bool
}

// NewConnection creates a new WebSocket connection
func NewConnection(id string, conn *websocket.Conn, manager *Manager, onMessage func(HubMessage), onDisconnect func(*Connection)) *Connection {
	return &Connection{
		ID:           id,
		ConnType:     ConnectionTypePlayer,
		Conn:         conn,
		Send:         make(chan dto.WebSocketMessage, 256),
		onMessage:    onMessage,
		onDisconnect: onDisconnect,
		manager:      manager,
		logger:       logger.Get(),
		Done:         make(chan struct{}),
	}
}

// SetPlayer associates this connection with a player
func (c *Connection) SetPlayer(playerID, gameID string) {
	c.mu.Lock()
	c.PlayerID = playerID
	c.GameID = gameID
	c.ConnType = ConnectionTypePlayer
	c.mu.Unlock()

	if c.manager != nil && gameID != "" {
		c.manager.AddToGame(c, gameID)
	}
}

// SetSpectator associates this connection with a spectator.
func (c *Connection) SetSpectator(spectatorID, gameID string) {
	c.mu.Lock()
	c.SpectatorID = spectatorID
	c.GameID = gameID
	c.ConnType = ConnectionTypeSpectator
	c.mu.Unlock()

	if c.manager != nil && gameID != "" {
		c.manager.AddToGame(c, gameID)
	}
}

// IsSpectator returns true if this connection is a spectator.
func (c *Connection) IsSpectator() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ConnType == ConnectionTypeSpectator
}

// GetPlayer returns the player and game IDs for this connection
func (c *Connection) GetPlayer() (playerID, gameID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PlayerID, c.GameID
}

// CloseSend closes the send channel
func (c *Connection) CloseSend() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.sendClosed {
		close(c.Send)
		c.sendClosed = true
	}
}

// Close closes the connection and signals all associated goroutines
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.Done)
		if err := c.Conn.Close(); err != nil {
			c.logger.Debug("Best-effort connection close", zap.Error(err), zap.String("connection_id", c.ID))
		}
	})
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Connection) ReadPump() {
	defer func() {
		if c.onDisconnect != nil {
			c.onDisconnect(c)
		}
		c.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Warn("Failed to set initial read deadline", zap.Error(err), zap.String("connection_id", c.ID))
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Warn("Failed to set read deadline in pong handler", zap.Error(err), zap.String("connection_id", c.ID))
		}
		return nil
	})

	c.logger.Debug("Starting ReadPump for connection", zap.String("connection_id", c.ID))

	for {
		select {
		case <-c.Done:
			return
		default:
			var message dto.WebSocketMessage
			if err := c.Conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					c.logger.Error("WebSocket read error", zap.Error(err), zap.String("connection_id", c.ID))
				}
				return
			}

			c.logger.Debug("Received WebSocket message",
				zap.String("connection_id", c.ID),
				zap.String("message_type", string(message.Type)))

			// Send message to hub for processing via callback
			if c.onMessage != nil {
				c.onMessage(HubMessage{Connection: c, Message: message})
			}
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			c.logger.Debug("Best-effort connection close in WritePump", zap.Error(err), zap.String("connection_id", c.ID))
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.logger.Warn("Failed to set write deadline", zap.Error(err), zap.String("connection_id", c.ID))
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.logger.Debug("Best-effort close message write", zap.Error(err), zap.String("connection_id", c.ID))
				}
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Error("WebSocket write error", zap.Error(err), zap.String("connection_id", c.ID))
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.logger.Warn("Failed to set write deadline for ping", zap.Error(err), zap.String("connection_id", c.ID))
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.Done:
			return
		}
	}
}

// SendMessage sends a message to this connection
func (c *Connection) SendMessage(message dto.WebSocketMessage) {
	c.mu.RLock()
	sendClosed := c.sendClosed
	c.mu.RUnlock()

	if sendClosed {
		c.logger.Debug("Attempted to send message to closed connection", zap.String("connection_id", c.ID))
		return
	}

	select {
	case c.Send <- message:
		c.logger.Debug("Message queued for client",
			zap.String("connection_id", c.ID),
			zap.String("message_type", string(message.Type)))
	case <-c.Done:
		c.logger.Debug("Connection closing, message not sent", zap.String("connection_id", c.ID))
	default:
		c.logger.Warn("Message channel full, dropping message",
			zap.String("connection_id", c.ID),
			zap.String("message_type", string(message.Type)))
	}
}
