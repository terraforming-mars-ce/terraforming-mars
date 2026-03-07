package game

import (
	"context"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// AddBotHandler handles add bot requests
type AddBotHandler struct {
	addBotAction *gameaction.AddBotAction
	broadcaster  Broadcaster
	logger       *zap.Logger
}

// NewAddBotHandler creates a new add bot handler
func NewAddBotHandler(addBotAction *gameaction.AddBotAction, broadcaster Broadcaster) *AddBotHandler {
	return &AddBotHandler{
		addBotAction: addBotAction,
		broadcaster:  broadcaster,
		logger:       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *AddBotHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing add bot request")

	_, gameID := connection.GetPlayer()
	if gameID == "" {
		if payloadMap, ok := message.Payload.(map[string]interface{}); ok {
			gameID, _ = payloadMap["gameId"].(string)
		}
	}

	if gameID == "" {
		log.Error("Missing gameId")
		h.sendError(connection, "Missing gameId")
		return
	}

	var botName string
	var difficulty string
	var speed string
	if payloadMap, ok := message.Payload.(map[string]interface{}); ok {
		botName, _ = payloadMap["botName"].(string)
		difficulty, _ = payloadMap["difficulty"].(string)
		speed, _ = payloadMap["speed"].(string)
	}

	result, err := h.addBotAction.Execute(ctx, gameID, botName, difficulty, speed)
	if err != nil {
		log.Error("Failed to add bot", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Bot added", zap.String("bot_id", result.PlayerID))

	h.broadcaster.BroadcastGameState(gameID, nil)
}

func (h *AddBotHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
