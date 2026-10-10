package game

import (
	"context"
	"log/slog"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// AddBotHandler handles add bot requests
type AddBotHandler struct {
	addBotAction *gameaction.AddBotAction
	broadcaster  Broadcaster
	logger       *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing add bot request")

	playerID, gameID := connection.GetPlayer()
	if gameID == "" || playerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	result, err := h.addBotAction.Execute(ctx, gameID, playerID)
	if err != nil {
		log.Warn("Failed to add bot", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Bot added", slog.String("bot_id", result.PlayerID))

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
