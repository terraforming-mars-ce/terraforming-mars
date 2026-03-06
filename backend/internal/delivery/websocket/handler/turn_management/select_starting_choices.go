package turn_management

import (
	"context"

	turnaction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// SelectStartingChoicesHandler handles combined starting selection requests
type SelectStartingChoicesHandler struct {
	action      *turnaction.SelectStartingChoicesAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSelectStartingChoicesHandler creates a new select starting choices handler
func NewSelectStartingChoicesHandler(action *turnaction.SelectStartingChoicesAction, broadcaster Broadcaster) *SelectStartingChoicesHandler {
	return &SelectStartingChoicesHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectStartingChoicesHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing select starting choices request")

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

	var preludeIDs []string
	if preludeIDsInterface, ok := payloadMap["preludeIds"].([]interface{}); ok {
		preludeIDs = make([]string, len(preludeIDsInterface))
		for i, preludeID := range preludeIDsInterface {
			if preludeIDStr, ok := preludeID.(string); ok {
				preludeIDs[i] = preludeIDStr
			}
		}
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

	log.Debug("Parsed select starting choices request",
		zap.String("corporation_id", corporationID),
		zap.Strings("prelude_ids", preludeIDs),
		zap.Strings("card_ids", cardIDs))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, corporationID, preludeIDs, cardIDs)
	if err != nil {
		log.Error("Failed to execute select starting choices action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Starting choices selected")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-starting-choices",
			"success": true,
		},
	}
}

func (h *SelectStartingChoicesHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
