package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmCardDiscardHandler handles confirm card discard requests
type ConfirmCardDiscardHandler struct {
	action      *confirmaction.ConfirmCardDiscardAction
	broadcaster Broadcaster
	logger      *zap.Logger
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
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
		zap.Strings("cards_to_discard", cardsToDiscard))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardsToDiscard)
	if err != nil {
		log.Error("Failed to execute confirm card discard action", zap.Error(err))
		h.sendError(connection, err.Error())
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
