package connection

import (
	"context"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// PlayerReconnectedHandler handles player reconnection requests
type PlayerReconnectedHandler struct {
	action *connaction.PlayerReconnectedAction
	logger *zap.Logger
}

// NewPlayerReconnectedHandler creates a new player reconnected handler
func NewPlayerReconnectedHandler(action *connaction.PlayerReconnectedAction) *PlayerReconnectedHandler {
	return &PlayerReconnectedHandler{
		action: action,
		logger: logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *PlayerReconnectedHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🔗 Processing player reconnected request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute player reconnected action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Player reconnected action completed successfully")

	response := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerReconnected,
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"playerId": connection.PlayerID,
			"success":  true,
		},
	}

	connection.Send <- response
}

func (h *PlayerReconnectedHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
