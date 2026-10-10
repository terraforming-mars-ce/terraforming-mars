package core

import (
	"context"
	"log/slog"
	"openmars/internal/delivery/dto"
	"openmars/internal/logger"
)

// MessageHandler defines the interface for handling different message types
type MessageHandler interface {
	HandleMessage(ctx context.Context, connection *Connection, message dto.WebSocketMessage)
}

// HubMessage represents a message to be processed by the hub
type HubMessage struct {
	Connection *Connection
	Message    dto.WebSocketMessage
}

// EventHandler interface for handling domain events
type EventHandler interface {
}

// hubJob is work from outside the hub, such as a bot, that must run on the hub goroutine.
type hubJob struct {
	ctx  context.Context
	fn   func()
	done chan struct{}
}

// Hub manages WebSocket connections and message routing. All game reads and writes
// happen on its single goroutine, so actions never run concurrently.
type Hub struct {
	Register   chan *Connection
	Unregister chan *Connection
	Messages   chan HubMessage
	jobs       chan hubJob

	manager  *Manager
	logger   *slog.Logger
	handlers map[dto.MessageType]MessageHandler
	guard    func(context.Context, *Connection, dto.WebSocketMessage) error
}

// NewHub creates a new WebSocket hub with clean architecture
func NewHub() *Hub {
	manager := NewManager()

	return &Hub{
		Register:   make(chan *Connection),
		Unregister: make(chan *Connection),
		Messages:   make(chan HubMessage),
		jobs:       make(chan hubJob),
		manager:    manager,
		logger:     logger.Get(),
		handlers:   make(map[dto.MessageType]MessageHandler),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Debug("Starting WebSocket hub")
	h.logger.Debug("WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Debug("WebSocket hub shutting down")
			h.manager.CloseAllConnections()
			return

		case connection := <-h.Register:
			h.manager.RegisterConnection(connection)
			// Session registration will happen when first message is received

		case connection := <-h.Unregister:
			playerID, spectatorID, gameID, connType, shouldBroadcast := h.manager.UnregisterConnection(connection)

			if shouldBroadcast {
				var disconnectMessage dto.WebSocketMessage

				if connType == ConnectionTypeSpectator {
					disconnectMessage = dto.WebSocketMessage{
						Type:   dto.MessageTypeSpectatorDisconnected,
						GameID: gameID,
						Payload: dto.SpectatorDisconnectedPayload{
							SpectatorID: spectatorID,
							GameID:      gameID,
						},
					}
				} else {
					disconnectMessage = dto.WebSocketMessage{
						Type:   dto.MessageTypePlayerDisconnected,
						GameID: gameID,
						Payload: dto.PlayerDisconnectedPayload{
							PlayerID: playerID,
							GameID:   gameID,
						},
					}
				}

				hubMessage := HubMessage{
					Connection: connection,
					Message:    disconnectMessage,
				}

				h.routeMessage(ctx, hubMessage)
			}

		case hubMessage := <-h.Messages:
			// Route message to appropriate handler
			h.routeMessage(ctx, hubMessage)

		case job := <-h.jobs:
			h.runJob(job)
		}
	}
}

// RegisterHandler registers a message handler for a specific message type
func (h *Hub) RegisterHandler(messageType dto.MessageType, handler MessageHandler) {
	h.handlers[messageType] = handler
}

// GetManager returns the connection manager
func (h *Hub) GetManager() *Manager {
	return h.manager
}

// SendToPlayer sends a message to a specific player via their connection
func (h *Hub) SendToPlayer(gameID, playerID string, message dto.WebSocketMessage) error {
	connection := h.manager.GetConnectionByPlayerID(gameID, playerID)
	if connection == nil {
		h.logger.Debug("No connection found for player",
			slog.String("game_id", gameID),
			slog.String("player_id", playerID))
		return nil // Don't error, just skip sending (player might be disconnected)
	}

	connection.SendMessage(message)
	h.logger.Debug("Message sent to player via Hub",
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("message_type", string(message.Type)))

	return nil
}

// SendToSpectator sends a message to a specific spectator via their connection.
func (h *Hub) SendToSpectator(gameID, spectatorID string, message dto.WebSocketMessage) error {
	connection := h.manager.GetConnectionBySpectatorID(gameID, spectatorID)
	if connection == nil {
		return nil
	}

	connection.SendMessage(message)
	return nil
}

// RegisterConnectionWithGame registers a connection with a game after player ID is set
func (h *Hub) RegisterConnectionWithGame(connection *Connection, gameID string) {
	h.manager.AddToGame(connection, gameID)
	h.logger.Debug("Connection registered with game",
		slog.String("connection_id", connection.ID),
		slog.String("game_id", gameID),
		slog.String("player_id", connection.PlayerID))
}

// Do runs fn on the hub goroutine and waits for it, so callers outside the hub can read
// and change game state without racing player actions. It must not be called from the
// hub goroutine itself. fn is skipped if ctx ends before the hub reaches it.
func (h *Hub) Do(ctx context.Context, fn func()) error {
	job := hubJob{ctx: ctx, fn: fn, done: make(chan struct{})}
	select {
	case h.jobs <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-job.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) runJob(job hubJob) {
	defer close(job.done)
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Hub job panicked", slog.Any("panic", r))
		}
	}()
	if job.ctx.Err() == nil {
		job.fn()
	}
}

// routeMessage routes incoming messages to appropriate handlers
func (h *Hub) routeMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message

	h.logger.Debug("Routing WebSocket message",
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)))

	if h.guard != nil {
		if err := h.guard(ctx, connection, message); err != nil {
			h.sendError(connection, err.Error())
			return
		}
	}
	if handler, exists := h.handlers[message.Type]; exists {
		h.logger.Debug("Routing to registered message handler",
			slog.String("message_type", string(message.Type)))
		handler.HandleMessage(ctx, connection, message)
	} else {
		h.logger.Warn("Unknown message type",
			slog.String("message_type", string(message.Type)))
		h.sendError(connection, ErrUnknownMessageType)
	}
}

// sendError sends an error message to a connection
func (h *Hub) sendError(connection *Connection, errorMessage string) {
	_, gameID := connection.GetPlayer()

	message := dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: errorMessage,
		},
		GameID: gameID,
	}

	connection.SendMessage(message)
}

// Hub no longer provides SessionManager - they're now separate components

// ClearConnections closes all active connections and clears the connection state
func (h *Hub) ClearConnections() {
	h.manager.CloseAllConnections()
}

// Standard error messages for hub operations
const (
	ErrHandlerNotAvailable = "Handler not available"
	ErrUnknownMessageType  = "Unknown message type"
)

// SetMessageGuard checks session restrictions before routing a client command.
func (h *Hub) SetMessageGuard(guard func(context.Context, *Connection, dto.WebSocketMessage) error) {
	h.guard = guard
}
