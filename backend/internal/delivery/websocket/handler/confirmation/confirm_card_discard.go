package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmCardDiscardHandler handles confirm card discard requests
type ConfirmCardDiscardHandler struct {
	action      *confirmaction.ConfirmCardDiscardAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmCardDiscardHandler creates a new confirm card discard handler
func NewConfirmCardDiscardHandler(action *confirmaction.ConfirmCardDiscardAction, broadcaster Broadcaster) *ConfirmCardDiscardHandler {
	return &ConfirmCardDiscardHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmCardDiscardHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm card discard request")

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

	resolutionID, ok := payloadMap["resolutionId"].(string)
	if !ok || resolutionID == "" {
		h.sendError(connection, "Missing resolutionId")
		return
	}

	var cardsToDiscard []string
	if cardsInterface, ok := payloadMap["cardsToDiscard"].([]interface{}); ok {
		cardsToDiscard = make([]string, len(cardsInterface))
		for i, cardID := range cardsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				cardsToDiscard[i] = cardIDStr
			}
		}
	}

	log.Debug("Parsed confirm card discard request",
		slog.Any("cards_to_discard", cardsToDiscard))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, resolutionID, cardsToDiscard)
	if err != nil {
		log.Error("Failed to execute confirm card discard action", slog.Any("error", err))
		connection.Send <- dto.WebSocketMessage{Type: dto.MessageTypeError, Payload: map[string]interface{}{"error": err.Error(), "resolutionId": resolutionID}}
		return
	}

	log.Debug("Card discard confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-card-discard",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmCardDiscardHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
