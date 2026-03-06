package turn_management

import (
	"context"

	turnaction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmInitAdvanceHandler handles init phase advance confirmation requests
type ConfirmInitAdvanceHandler struct {
	action      *turnaction.ConfirmInitAdvanceAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmInitAdvanceHandler creates a new confirm init advance handler
func NewConfirmInitAdvanceHandler(action *turnaction.ConfirmInitAdvanceAction, broadcaster Broadcaster) *ConfirmInitAdvanceHandler {
	return &ConfirmInitAdvanceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmInitAdvanceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing init phase advance request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute init advance", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Init advance completed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-init-advance",
			"success": true,
		},
	}
}

func (h *ConfirmInitAdvanceHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
