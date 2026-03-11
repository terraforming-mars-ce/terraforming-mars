package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmColonyResourceHandler handles confirm colony resource placement requests
type ConfirmColonyResourceHandler struct {
	action      *confirmaction.ConfirmColonyResourceAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmColonyResourceHandler creates a new confirm colony resource handler
func NewConfirmColonyResourceHandler(action *confirmaction.ConfirmColonyResourceAction, broadcaster Broadcaster) *ConfirmColonyResourceHandler {
	return &ConfirmColonyResourceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmColonyResourceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm colony resource request")

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

	targetCardID, _ := payloadMap["cardId"].(string)

	log.Debug("Parsed confirm colony resource request",
		zap.String("target_card_id", targetCardID))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, targetCardID)
	if err != nil {
		log.Error("Failed to execute confirm colony resource action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Colony resource placement confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-colony-resource",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmColonyResourceHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
