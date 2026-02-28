package connection

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// RequestLogsHandler handles requests to resend all game logs
type RequestLogsHandler struct {
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewRequestLogsHandler creates a new request logs handler
func NewRequestLogsHandler(broadcaster Broadcaster) *RequestLogsHandler {
	return &RequestLogsHandler{
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *RequestLogsHandler) HandleMessage(_ context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("📋 Processing request-logs")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		connection.Send <- dto.WebSocketMessage{
			Type: dto.MessageTypeError,
			Payload: map[string]any{
				"error": "Not connected to a game",
			},
		}
		return
	}

	h.broadcaster.SendInitialLogs(connection.GameID, connection.PlayerID)
	log.Info("✅ Sent initial logs to player",
		zap.String("game_id", connection.GameID),
		zap.String("player_id", connection.PlayerID))
}
