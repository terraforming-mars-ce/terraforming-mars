package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmAwardFundHandler handles confirm award fund requests
type ConfirmAwardFundHandler struct {
	action      *confirmaction.ConfirmAwardFundAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmAwardFundHandler creates a new confirm award fund handler
func NewConfirmAwardFundHandler(action *confirmaction.ConfirmAwardFundAction, broadcaster Broadcaster) *ConfirmAwardFundHandler {
	return &ConfirmAwardFundHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmAwardFundHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm award fund request")

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

	awardType, ok := payloadMap["awardType"].(string)
	if !ok || awardType == "" {
		log.Error("Missing or invalid awardType in payload")
		h.sendError(connection, "Missing awardType")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, awardType)
	if err != nil {
		log.Error("Failed to execute confirm award fund action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Award fund confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-award-fund",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmAwardFundHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
