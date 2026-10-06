package confirmation

import (
	"context"
	"encoding/json"
	"log/slog"
	"terraforming-mars-backend/internal/game/shared"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
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

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
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
		h.sendError(connection, "Invalid payment")
		return
	}
	if paymentErr = json.Unmarshal(paymentBytes, &paymentEnvelope); paymentErr != nil {
		h.sendError(connection, "Invalid payment")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, selectedCardIDs, randomBuy, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute confirm production cards action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Production cards confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-production-cards",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmProductionCardsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
