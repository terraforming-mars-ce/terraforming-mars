package turn_management

import (
	"context"

	turnaction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SelectPreludeCardsHandler handles select prelude cards requests
type SelectPreludeCardsHandler struct {
	action      *turnaction.SelectPreludeCardsAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSelectPreludeCardsHandler creates a new select prelude cards handler
func NewSelectPreludeCardsHandler(action *turnaction.SelectPreludeCardsAction, broadcaster Broadcaster) *SelectPreludeCardsHandler {
	return &SelectPreludeCardsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectPreludeCardsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🃏 Processing select prelude cards request")

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

	var preludeIDs []string
	if preludeIDsInterface, ok := payloadMap["preludeIds"].([]interface{}); ok {
		preludeIDs = make([]string, len(preludeIDsInterface))
		for i, preludeID := range preludeIDsInterface {
			if preludeIDStr, ok := preludeID.(string); ok {
				preludeIDs[i] = preludeIDStr
			}
		}
	}

	log.Debug("Parsed select prelude cards request",
		zap.Strings("prelude_ids", preludeIDs))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, preludeIDs)
	if err != nil {
		log.Error("Failed to execute select prelude cards action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Select prelude cards action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-prelude-cards",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SelectPreludeCardsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
