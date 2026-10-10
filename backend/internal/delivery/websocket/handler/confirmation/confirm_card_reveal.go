package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmCardRevealHandler handles confirm card reveal requests
type ConfirmCardRevealHandler struct {
	action      *confirmaction.ConfirmCardRevealAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmCardRevealHandler creates a new confirm card reveal handler
func NewConfirmCardRevealHandler(action *confirmaction.ConfirmCardRevealAction, broadcaster Broadcaster) *ConfirmCardRevealHandler {
	return &ConfirmCardRevealHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmCardRevealHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm card reveal request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID())
	if err != nil {
		log.Error("Failed to execute confirm card reveal action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Card reveal confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}
