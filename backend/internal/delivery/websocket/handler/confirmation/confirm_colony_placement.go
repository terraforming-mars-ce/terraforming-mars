package confirmation

import (
	"context"

	confirmaction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmColonyPlacementHandler handles confirm colony placement requests
type ConfirmColonyPlacementHandler struct {
	action      *confirmaction.ConfirmColonyPlacementAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmColonyPlacementHandler creates a new confirm colony placement handler
func NewConfirmColonyPlacementHandler(action *confirmaction.ConfirmColonyPlacementAction, broadcaster Broadcaster) *ConfirmColonyPlacementHandler {
	return &ConfirmColonyPlacementHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmColonyPlacementHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm colony placement request")

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

	colonyID, _ := payloadMap["colonyId"].(string)
	if colonyID == "" {
		h.sendError(connection, "Missing colonyId in payload")
		return
	}

	log.Debug("Parsed confirm colony placement request",
		zap.String("colony_id", colonyID))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, colonyID)
	if err != nil {
		log.Error("Failed to execute confirm colony placement action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Colony placement confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-colony-placement",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmColonyPlacementHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
