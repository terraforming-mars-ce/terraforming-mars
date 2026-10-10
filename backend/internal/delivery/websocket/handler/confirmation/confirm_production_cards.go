package confirmation

import (
	"context"
	"encoding/json"
	"log/slog"
	"openmars/internal/game/shared"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmProductionCardsHandler handles confirm production cards requests
type ConfirmProductionCardsHandler struct {
	action      *confirmaction.ConfirmProductionCardsAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmProductionCardsHandler creates a new confirm production cards handler
func NewConfirmProductionCardsHandler(action *confirmaction.ConfirmProductionCardsAction, broadcaster Broadcaster) *ConfirmProductionCardsHandler {
	return &ConfirmProductionCardsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmProductionCardsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm production cards request")

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
	if cardIDsInterface, ok := payloadMap["cardIds"].([]interface{}); ok {
		selectedCardIDs = make([]string, len(cardIDsInterface))
		for i, cardID := range cardIDsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				selectedCardIDs[i] = cardIDStr
			}
		}
	}

	randomBuy, _ := payloadMap["randomBuy"].(bool)

	log.Debug("Parsed confirm production cards request",
		slog.Any("selected_card_ids", selectedCardIDs),
		slog.Bool("random_buy", randomBuy))

	var paymentEnvelope struct {
		Payment shared.Payment `json:"payment"`
	}
	paymentBytes, paymentErr := json.Marshal(message.Payload)
	if paymentErr != nil {
		connection.SendError(message.Type, "Invalid payment")
		return
	}
	if paymentErr = json.Unmarshal(paymentBytes, &paymentEnvelope); paymentErr != nil {
		connection.SendError(message.Type, "Invalid payment")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), selectedCardIDs, randomBuy, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute confirm production cards action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Production cards confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
