package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmStealTargetHandler handles confirm steal target requests
type ConfirmStealTargetHandler struct {
	action      *confirmaction.ConfirmStealTargetAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmStealTargetHandler creates a new confirm steal target handler
func NewConfirmStealTargetHandler(action *confirmaction.ConfirmStealTargetAction, broadcaster Broadcaster) *ConfirmStealTargetHandler {
	return &ConfirmStealTargetHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmStealTargetHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm steal target request")

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

	targetPlayerId, _ := payloadMap["targetPlayerId"].(string)

	log.Debug("Parsed confirm steal target request",
		zap.String("target_player_id", targetPlayerId))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, targetPlayerId)
	if err != nil {
		log.Error("Failed to execute confirm steal target action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Steal target confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-steal-target",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmStealTargetHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
