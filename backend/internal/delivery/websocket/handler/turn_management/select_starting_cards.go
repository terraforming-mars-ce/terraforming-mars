package turn_management

import (
	"context"

	turnaction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SelectStartingCardsHandler handles select starting cards requests
type SelectStartingCardsHandler struct {
	action      *turnaction.SelectStartingCardsAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewSelectStartingCardsHandler creates a new select starting cards handler
func NewSelectStartingCardsHandler(action *turnaction.SelectStartingCardsAction, broadcaster Broadcaster) *SelectStartingCardsHandler {
	return &SelectStartingCardsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectStartingCardsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🃏 Processing select starting cards request")

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

	var cardIDs []string
	if cardIDsInterface, ok := payloadMap["cardIds"].([]interface{}); ok {
		cardIDs = make([]string, len(cardIDsInterface))
		for i, cardID := range cardIDsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				cardIDs[i] = cardIDStr
			}
		}
	}

	corporationID, _ := payloadMap["corporationId"].(string)

	log.Debug("Parsed select starting cards request",
		zap.Strings("card_ids", cardIDs),
		zap.String("corporation_id", corporationID))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardIDs, corporationID)
	if err != nil {
		log.Error("Failed to execute select starting cards action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Select starting cards action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-starting-cards",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SelectStartingCardsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
