package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmSellPatentsHandler handles confirm sell patents requests
type ConfirmSellPatentsHandler struct {
	action      *confirmaction.ConfirmSellPatentsAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmSellPatentsHandler creates a new confirm sell patents handler
func NewConfirmSellPatentsHandler(action *confirmaction.ConfirmSellPatentsAction, broadcaster Broadcaster) *ConfirmSellPatentsHandler {
	return &ConfirmSellPatentsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmSellPatentsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm sell patents request")

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

	var selectedCardIDs []string
	if cardIDsInterface, ok := payloadMap["selectedCardIds"].([]interface{}); ok {
		selectedCardIDs = make([]string, len(cardIDsInterface))
		for i, cardID := range cardIDsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				selectedCardIDs[i] = cardIDStr
			}
		}
	}

	log.Debug("Parsed confirm sell patents request",
		slog.Any("selected_card_ids", selectedCardIDs))

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), selectedCardIDs)
	if err != nil {
		log.Error("Failed to execute confirm sell patents action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Sell patents confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
