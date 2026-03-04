package core

import (
	"context"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
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

// Hub manages WebSocket connections and message routing
type Hub struct {
	Register   chan *Connection
	Unregister chan *Connection
	Messages   chan HubMessage

	manager  *Manager
	logger   *zap.Logger
	handlers map[dto.MessageType]MessageHandler
}

// NewHub creates a new WebSocket hub with clean architecture
func NewHub() *Hub {
	manager := NewManager()

	return &Hub{
		Register:   make(chan *Connection),
		Unregister: make(chan *Connection),
		Messages:   make(chan HubMessage),
		manager:    manager,
		logger:     logger.Get(),
		handlers:   make(map[dto.MessageType]MessageHandler),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("🚀 Starting WebSocket hub")
	h.logger.Info("✅ WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("🛑 WebSocket hub shutting down")
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
		h.logger.Debug("❌ No connection found for player",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		return nil // Don't error, just skip sending (player might be disconnected)
	}

	connection.SendMessage(message)
	h.logger.Debug("💬 Message sent to player via Hub",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("message_type", string(message.Type)))

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
	h.logger.Debug("🎯 Connection registered with game",
		zap.String("connection_id", connection.ID),
		zap.String("game_id", gameID),
		zap.String("player_id", connection.PlayerID))
}

// routeMessage routes incoming messages to appropriate handlers
func (h *Hub) routeMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message

	h.logger.Info("🔄 Routing WebSocket message",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))

	if handler, exists := h.handlers[message.Type]; exists {
		h.logger.Debug("🎯 Routing to registered message handler",
			zap.String("message_type", string(message.Type)))
		handler.HandleMessage(ctx, connection, message)
	} else {
		h.logger.Warn("❓ Unknown message type",
			zap.String("message_type", string(message.Type)))
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
