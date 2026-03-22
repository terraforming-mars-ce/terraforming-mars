package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmFreeTradeHandler handles confirm free trade requests
type ConfirmFreeTradeHandler struct {
	action      *confirmaction.ConfirmFreeTradeAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmFreeTradeHandler creates a new confirm free trade handler
func NewConfirmFreeTradeHandler(action *confirmaction.ConfirmFreeTradeAction, broadcaster Broadcaster) *ConfirmFreeTradeHandler {
	return &ConfirmFreeTradeHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmFreeTradeHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm free trade request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		h.sendError(connection, "Invalid payload format")
		return
	}

	colonyID, _ := payloadMap["colonyId"].(string)
	if colonyID == "" {
		h.sendError(connection, "Missing colonyId in payload")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, colonyID)
	if err != nil {
		log.Error("Failed to execute confirm free trade action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Free trade confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-free-trade",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmFreeTradeHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
