package card

import (
	"context"
	"encoding/json"
	"log/slog"

	cardaction "openmars/internal/action/card"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// PlayCardHandler handles play card requests
type PlayCardHandler struct {
	action      *cardaction.PlayCardAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewPlayCardHandler creates a new play card handler
func NewPlayCardHandler(action *cardaction.PlayCardAction, broadcaster Broadcaster) *PlayCardHandler {
	return &PlayCardHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *PlayCardHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing play card request")
	payload, payloadOK := message.Payload.(map[string]interface{})
	cardID, _ := payload["cardId"].(string)

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendErrorPayload(dto.ErrorPayload{Message: "Not connected to a game", RequestType: message.Type, CardID: cardID})
		return
	}

	if !payloadOK {
		log.Error("Invalid payload format")
		connection.SendErrorPayload(dto.ErrorPayload{Message: "Invalid payload format", RequestType: message.Type, CardID: cardID})
		return
	}

	if cardID == "" {
		log.Error("Missing or invalid cardId")
		connection.SendErrorPayload(dto.ErrorPayload{Message: "Missing cardId", RequestType: message.Type, CardID: cardID})
		return
	}

	var payment dto.PaymentDto
	raw, err := json.Marshal(payload["payment"])
	if err != nil {
		connection.SendErrorPayload(dto.ErrorPayload{Message: "Invalid payment", RequestType: message.Type, CardID: cardID})
		return
	}
	if err = json.Unmarshal(raw, &payment); err != nil {
		connection.SendErrorPayload(dto.ErrorPayload{Message: "Invalid payment", RequestType: message.Type, CardID: cardID})
		return
	}

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
				connection.SendErrorPayload(dto.ErrorPayload{Message: "Invalid storage input source", RequestType: message.Type, CardID: cardID})
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

	var selectedAmount *int
	if saFloat, ok := payload["selectedAmount"].(float64); ok {
		if saFloat < 0 || saFloat > 2147483647 || saFloat != float64(int(saFloat)) {
			connection.SendErrorPayload(dto.ErrorPayload{Message: "Invalid selected amount", RequestType: message.Type, CardID: cardID})
			return
		}
		sa := int(saFloat)
		selectedAmount = &sa
	}

	if choiceIndex != nil {
		log.Debug("Choice index extracted", slog.Int("choice_index", *choiceIndex))
	}
	if len(cardStorageTargets) > 0 {
		log.Debug("Card storage targets extracted", slog.Any("card_storage_targets", cardStorageTargets))
	}
	if targetPlayerID != nil {
		log.Debug("Target player extracted", slog.String("target_player_id", *targetPlayerID))
	}

	err = h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), cardID, dto.ToPayment(payment), choiceIndex, cardStorageTargets, targetPlayerID, selectedAmount, cardStorageSources)
	if err != nil {
		log.Error("Failed to execute play card action", slog.Any("error", err))
		connection.SendErrorPayload(dto.ErrorPayload{Message: err.Error(), RequestType: message.Type, CardID: cardID})
		return
	}

	log.Debug("Play card completed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
