package turn_management

import (
	"context"

	turnaction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SelectCorporationHandler handles select corporation requests
type SelectCorporationHandler struct {
	action      *turnaction.SelectCorporationAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSelectCorporationHandler creates a new select corporation handler
func NewSelectCorporationHandler(action *turnaction.SelectCorporationAction, broadcaster Broadcaster) *SelectCorporationHandler {
	return &SelectCorporationHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectCorporationHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🏢 Processing select corporation request")

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

	corporationID, _ := payloadMap["corporationId"].(string)
	if corporationID == "" {
		log.Error("Missing corporationId")
		h.sendError(connection, "Missing corporationId")
		return
	}

	log.Debug("Parsed select corporation request",
		zap.String("corporation_id", corporationID))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, corporationID)
	if err != nil {
		log.Error("Failed to execute select corporation action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Select corporation action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-corporation",
			"success": true,
		},
	}
}

func (h *SelectCorporationHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
