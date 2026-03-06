package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmSellPatentsHandler handles confirm sell patents requests
type ConfirmSellPatentsHandler struct {
	action      *confirmaction.ConfirmSellPatentsAction
	broadcaster Broadcaster
	logger      *zap.Logger
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm sell patents request")

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
	if cardIDsInterface, ok := payloadMap["selectedCardIds"].([]interface{}); ok {
		selectedCardIDs = make([]string, len(cardIDsInterface))
		for i, cardID := range cardIDsInterface {
			if cardIDStr, ok := cardID.(string); ok {
				selectedCardIDs[i] = cardIDStr
			}
		}
	}

	log.Debug("Parsed confirm sell patents request",
		zap.Strings("selected_card_ids", selectedCardIDs))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, selectedCardIDs)
	if err != nil {
		log.Error("Failed to execute confirm sell patents action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Sell patents confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-sell-patents",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmSellPatentsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
