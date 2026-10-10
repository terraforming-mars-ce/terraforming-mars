package core

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"openmars/internal/delivery/dto"
	"openmars/internal/logger"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 64 * 1024

	// Messages queued for a client that is not reading. A client this far behind is
	// disconnected, so it reconnects and gets the full state instead of silently
	// missing updates.
	sendBufferSize = 256
)

// Identity is who a connection acts as: a player or a spectator of one game, or nobody yet.
type Identity struct {
	GameID      string
	PlayerID    string
	SpectatorID string
}

// IsPlayer reports whether the connection acts as a player in a game.
func (id Identity) IsPlayer() bool {
	return id.GameID != "" && id.PlayerID != ""
}

// IsSpectator reports whether the connection watches a game.
func (id Identity) IsSpectator() bool {
	return id.GameID != "" && id.SpectatorID != ""
}

// Connection is one client WebSocket connection.
type Connection struct {
	ID string

	conn *websocket.Conn
	hub  *Hub
	send chan dto.WebSocketMessage

	// flush asks the write pump to send what is queued, then close the connection.
	flush     chan struct{}
	flushOnce sync.Once
	done      chan struct{}
	closeOnce sync.Once

	mu       sync.RWMutex
	identity Identity

	logger *slog.Logger
}

func newConnection(id string, conn *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:     id,
		conn:   conn,
		hub:    hub,
		send:   make(chan dto.WebSocketMessage, sendBufferSize),
		flush:  make(chan struct{}),
		done:   make(chan struct{}),
		logger: logger.Get(),
	}
}

// Identity returns who the connection currently acts as.
func (c *Connection) Identity() Identity {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.identity
}

// GameID returns the game the connection belongs to, or "" when it has not joined one.
func (c *Connection) GameID() string {
	return c.Identity().GameID
}

// PlayerID returns the player the connection acts as, or "" when it is not a player.
func (c *Connection) PlayerID() string {
	return c.Identity().PlayerID
}

// SpectatorID returns the spectator the connection acts as, or "" when it is not one.
func (c *Connection) SpectatorID() string {
	return c.Identity().SpectatorID
}

// IsSpectator reports whether the connection watches a game.
func (c *Connection) IsSpectator() bool {
	return c.Identity().IsSpectator()
}

// BindPlayer makes the connection act as a player. If it acted as someone else before,
// that identity is released. Must run on the hub goroutine.
func (c *Connection) BindPlayer(gameID, playerID string) {
	c.hub.bind(c, Identity{GameID: gameID, PlayerID: playerID})
}

// BindSpectator makes the connection watch a game. If it acted as someone else before,
// that identity is released. Must run on the hub goroutine.
func (c *Connection) BindSpectator(gameID, spectatorID string) {
	c.hub.bind(c, Identity{GameID: gameID, SpectatorID: spectatorID})
}

func (c *Connection) setIdentity(id Identity) Identity {
	c.mu.Lock()
	defer c.mu.Unlock()
	previous := c.identity
	c.identity = id
	return previous
}

// Send queues a message for the client without blocking. A client whose queue is full is
// not keeping up and is disconnected.
func (c *Connection) Send(message dto.WebSocketMessage) {
	select {
	case <-c.done:
		return
	default:
	}
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Client is not reading, disconnecting it",
			slog.String("connection_id", c.ID),
			slog.String("message_type", string(message.Type)))
		c.Close()
	}
}

// SendError tells the client its request failed.
func (c *Connection) SendError(requestType dto.MessageType, message string) {
	c.SendErrorPayload(dto.ErrorPayload{Message: message, RequestType: requestType})
}

// SendErrorPayload tells the client its request failed, with request-specific context.
func (c *Connection) SendErrorPayload(payload dto.ErrorPayload) {
	c.Send(dto.WebSocketMessage{Type: dto.MessageTypeError, GameID: c.GameID(), Payload: payload})
}

// CloseAfterFlush delivers everything already queued, then closes the connection.
func (c *Connection) CloseAfterFlush() {
	c.flushOnce.Do(func() { close(c.flush) })
}

// Close closes the connection immediately.
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		if err := c.conn.Close(); err != nil {
			c.logger.Debug("Best-effort connection close", slog.Any("error", err), slog.String("connection_id", c.ID))
		}
	})
}

// readPump hands every incoming message to the hub until the connection ends, then tells
// the hub the connection is gone.
func (c *Connection) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Debug("Failed to set read deadline", slog.Any("error", err), slog.String("connection_id", c.ID))
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var message dto.WebSocketMessage
		if err := c.conn.ReadJSON(&message); err != nil {
			if isMalformedJSON(err) {
				c.SendError("", "Malformed message: "+err.Error())
				continue
			}
			c.logger.Debug("WebSocket read ended", slog.Any("error", err), slog.String("connection_id", c.ID))
			return
		}
		if !c.hub.submit(HubMessage{Connection: c, Message: message}) {
			return
		}
	}
}

// writePump writes queued messages and keep-alive pings to the client.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message := <-c.send:
			if !c.write(message) {
				return
			}
		case <-c.flush:
			for {
				select {
				case message := <-c.send:
					if !c.write(message) {
						return
					}
				default:
					_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
					_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					return
				}
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

// isMalformedJSON reports whether a read failed only because the frame was not a valid
// message. The frame is consumed, so the connection can keep reading.
func isMalformedJSON(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.Is(err, io.ErrUnexpectedEOF)
}

func (c *Connection) write(message dto.WebSocketMessage) bool {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteJSON(message); err != nil {
		c.logger.Debug("WebSocket write failed", slog.Any("error", err), slog.String("connection_id", c.ID))
		return false
	}
	return true
}

// Done is closed when the connection closes.
func (c *Connection) Done() <-chan struct{} {
	return c.done
}
