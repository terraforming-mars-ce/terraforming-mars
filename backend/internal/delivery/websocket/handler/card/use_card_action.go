package card

import (
	"context"
	"encoding/json"
	"log/slog"

	cardaction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// UseCardActionHandler handles card action execution requests
type UseCardActionHandler struct {
	action      *cardaction.UseCardActionAction
	broadcaster Broadcaster
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing use card action request")

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

	var cardStorageSources []string
	if raw, ok := payload["cardStorageSources"].([]interface{}); ok {
		for _, v := range raw {
			id, ok := v.(string)
			if !ok {
				h.sendError(connection, "Invalid storage input source")
				return
			}
			cardStorageSources = append(cardStorageSources, id)
		}
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
		if saFloat < 0 || saFloat > 2147483647 || saFloat != float64(int(saFloat)) {
			h.sendError(connection, "Invalid selected amount")
			return
		}
		sa := int(saFloat)
		selectedAmount = &sa
	}

	var actionPayment *shared.Payment
	if value, ok := payload["payment"]; ok {
		raw, err := json.Marshal(value)
		if err != nil {
			h.sendError(connection, "Invalid payment")
			return
		}
		var payment shared.Payment
		if err = json.Unmarshal(raw, &payment); err != nil {
			h.sendError(connection, "Invalid payment")
			return
		}
		actionPayment = &payment
	}

	var reuseSourceCardID *string
	if rsci, ok := payload["reuseSourceCardId"].(string); ok && rsci != "" {
		reuseSourceCardID = &rsci
	}

	log = log.With(
		slog.String("card_id", cardID),
		slog.Int("behavior_index", behaviorIndex),
	)
	if choiceIndex != nil {
		log = log.With(slog.Int("choice_index", *choiceIndex))
	}
	if len(cardStorageTargets) > 0 {
		log = log.With(slog.Any("card_storage_targets", cardStorageTargets))
	}
	if targetPlayerID != nil {
		log = log.With(slog.String("target_player_id", *targetPlayerID))
	}
	if stealSourceCardID != nil {
		log = log.With(slog.String("source_card_for_input", *stealSourceCardID))
	}
	if reuseSourceCardID != nil {
		log = log.With(slog.String("reuse_source_card_id", *reuseSourceCardID))
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardID, behaviorIndex, choiceIndex, cardStorageTargets, targetPlayerID, stealSourceCardID, selectedAmount, actionPayment, reuseSourceCardID, cardStorageSources)
	if err != nil {
		log.Error("Failed to execute use card action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Card action completed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

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
