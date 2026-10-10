package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmColonyResourceHandler handles confirm colony resource placement requests
type ConfirmColonyResourceHandler struct {
	action      *confirmaction.ConfirmColonyResourceAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmColonyResourceHandler creates a new confirm colony resource handler
func NewConfirmColonyResourceHandler(action *confirmaction.ConfirmColonyResourceAction, broadcaster Broadcaster) *ConfirmColonyResourceHandler {
	return &ConfirmColonyResourceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmColonyResourceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm colony resource request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	targetCardID, _ := payloadMap["cardId"].(string)

	log.Debug("Parsed confirm colony resource request",
		slog.String("target_card_id", targetCardID))

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), targetCardID)
	if err != nil {
		log.Error("Failed to execute confirm colony resource action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Colony resource placement confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}
