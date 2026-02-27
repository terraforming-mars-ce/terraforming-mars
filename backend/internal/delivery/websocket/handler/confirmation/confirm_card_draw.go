package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmCardDrawHandler handles confirm card draw requests
type ConfirmCardDrawHandler struct {
	action      *confirmaction.ConfirmCardDrawAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewConfirmCardDrawHandler creates a new confirm card draw handler
func NewConfirmCardDrawHandler(action *confirmaction.ConfirmCardDrawAction, broadcaster Broadcaster) *ConfirmCardDrawHandler {
	return &ConfirmCardDrawHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmCardDrawHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎴 Processing confirm card draw request")

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

	var cardsToTake []string
	if cardsInterface, ok := payloadMap["cardsToTake"].([]interface{}); ok {
		cardsToTake = make([]string, len(cardsInterface))
		for i, cardID := range cardsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				cardsToTake[i] = cardIDStr
			}
		}
	}

	var cardsToBuy []string
	if cardsInterface, ok := payloadMap["cardsToBuy"].([]interface{}); ok {
		cardsToBuy = make([]string, len(cardsInterface))
		for i, cardID := range cardsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				cardsToBuy[i] = cardIDStr
			}
		}
	}

	log.Debug("Parsed confirm card draw request",
		zap.Strings("cards_to_take", cardsToTake),
		zap.Strings("cards_to_buy", cardsToBuy))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardsToTake, cardsToBuy)
	if err != nil {
		log.Error("Failed to execute confirm card draw action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Confirm card draw action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-card-draw",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmCardDrawHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
