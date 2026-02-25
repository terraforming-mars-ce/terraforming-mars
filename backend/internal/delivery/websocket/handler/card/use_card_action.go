package card

import (
	"context"

	cardaction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// UseCardActionHandler handles card action execution requests
type UseCardActionHandler struct {
	action      *cardaction.UseCardActionAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewUseCardActionHandler creates a new use card action handler
func NewUseCardActionHandler(action *cardaction.UseCardActionAction, broadcaster Broadcaster) *UseCardActionHandler {
	return &UseCardActionHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *UseCardActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎯 Processing use card action request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	cardID, ok := payload["cardId"].(string)
	if !ok || cardID == "" {
		log.Error("Missing or invalid cardId")
		h.sendError(connection, "Missing cardId")
		return
	}

	behaviorIndexFloat, ok := payload["behaviorIndex"].(float64)
	if !ok {
		log.Error("Missing or invalid behaviorIndex")
		h.sendError(connection, "Missing behaviorIndex")
		return
	}
	behaviorIndex := int(behaviorIndexFloat)

	var choiceIndex *int
	if choiceIndexFloat, ok := payload["choiceIndex"].(float64); ok {
		idx := int(choiceIndexFloat)
		choiceIndex = &idx
	}

	var cardStorageTargets []string
	if targetsRaw, ok := payload["cardStorageTargets"].([]interface{}); ok {
		for _, t := range targetsRaw {
			if s, ok := t.(string); ok {
				cardStorageTargets = append(cardStorageTargets, s)
			}
		}
	}

	var targetPlayerID *string
	if tpID, ok := payload["targetPlayerId"].(string); ok && tpID != "" {
		targetPlayerID = &tpID
	}

	var stealSourceCardID *string
	if scfi, ok := payload["sourceCardForInput"].(string); ok && scfi != "" {
		stealSourceCardID = &scfi
	}

	var selectedAmount *int
	if saFloat, ok := payload["selectedAmount"].(float64); ok {
		sa := int(saFloat)
		selectedAmount = &sa
	}

	var actionPayment *gamecards.CardPayment
	if paymentMap, ok := payload["payment"].(map[string]interface{}); ok {
		payment := &gamecards.CardPayment{}
		if credits, ok := paymentMap["credits"].(float64); ok {
			payment.Credits = int(credits)
		}
		if steel, ok := paymentMap["steel"].(float64); ok {
			payment.Steel = int(steel)
		}
		if titanium, ok := paymentMap["titanium"].(float64); ok {
			payment.Titanium = int(titanium)
		}
		actionPayment = payment
	}

	log = log.With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
	)
	if choiceIndex != nil {
		log = log.With(zap.Int("choice_index", *choiceIndex))
	}
	if len(cardStorageTargets) > 0 {
		log = log.With(zap.Strings("card_storage_targets", cardStorageTargets))
	}
	if targetPlayerID != nil {
		log = log.With(zap.String("target_player_id", *targetPlayerID))
	}
	if stealSourceCardID != nil {
		log = log.With(zap.String("source_card_for_input", *stealSourceCardID))
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardID, behaviorIndex, choiceIndex, cardStorageTargets, targetPlayerID, stealSourceCardID, selectedAmount, actionPayment)
	if err != nil {
		log.Error("Failed to execute use card action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Use card action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":        "card-action",
			"success":       true,
			"cardId":        cardID,
			"behaviorIndex": behaviorIndex,
		},
	}

	connection.Send <- response
}

func (h *UseCardActionHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
