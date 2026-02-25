package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmBehaviorChoiceHandler handles confirm behavior choice requests
type ConfirmBehaviorChoiceHandler struct {
	action      *confirmaction.ConfirmBehaviorChoiceAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmBehaviorChoiceHandler creates a new confirm behavior choice handler
func NewConfirmBehaviorChoiceHandler(action *confirmaction.ConfirmBehaviorChoiceAction, broadcaster Broadcaster) *ConfirmBehaviorChoiceHandler {
	return &ConfirmBehaviorChoiceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmBehaviorChoiceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🔀 Processing confirm behavior choice request")

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

	choiceIndexFloat, ok := payloadMap["choiceIndex"].(float64)
	if !ok {
		log.Error("Missing or invalid choiceIndex")
		h.sendError(connection, "Missing or invalid choiceIndex")
		return
	}
	choiceIndex := int(choiceIndexFloat)

	var cardStorageTargets []string
	if targetsRaw, ok := payloadMap["cardStorageTargets"].([]interface{}); ok {
		for _, t := range targetsRaw {
			if s, ok := t.(string); ok {
				cardStorageTargets = append(cardStorageTargets, s)
			}
		}
	}

	log.Debug("Parsed confirm behavior choice request",
		zap.Int("choice_index", choiceIndex))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, choiceIndex, cardStorageTargets)
	if err != nil {
		log.Error("Failed to execute confirm behavior choice action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Confirm behavior choice action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-behavior-choice",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmBehaviorChoiceHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
