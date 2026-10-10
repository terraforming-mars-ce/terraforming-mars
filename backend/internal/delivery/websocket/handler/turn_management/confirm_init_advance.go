package turn_management

import (
	"context"
	"log/slog"

	turnaction "openmars/internal/action/turn_management"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmInitAdvanceHandler handles init phase advance confirmation requests
type ConfirmInitAdvanceHandler struct {
	action      *turnaction.ConfirmInitAdvanceAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmInitAdvanceHandler creates a new confirm init advance handler
func NewConfirmInitAdvanceHandler(action *turnaction.ConfirmInitAdvanceAction, broadcaster Broadcaster) *ConfirmInitAdvanceHandler {
	return &ConfirmInitAdvanceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmInitAdvanceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing init phase advance request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID())
	if err != nil {
		log.Error("Failed to execute init advance", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Init advance completed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}
