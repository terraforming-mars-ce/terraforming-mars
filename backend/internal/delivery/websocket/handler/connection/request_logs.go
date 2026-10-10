package connection

import (
	"context"
	"log/slog"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// RequestLogsHandler handles requests to resend all game logs
type RequestLogsHandler struct {
	broadcaster Broadcaster
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing request-logs")

	if connection.GameID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	if connection.SpectatorID() != "" {
		h.broadcaster.SendInitialLogsToSpectator(connection.GameID(), connection.SpectatorID())
		log.Debug("Sent initial logs to spectator",
			slog.String("game_id", connection.GameID()),
			slog.String("spectator_id", connection.SpectatorID()))
		return
	}

	if connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	h.broadcaster.SendInitialLogs(connection.GameID(), connection.PlayerID())
	log.Debug("Sent initial logs to player",
		slog.String("game_id", connection.GameID()),
		slog.String("player_id", connection.PlayerID()))
}
