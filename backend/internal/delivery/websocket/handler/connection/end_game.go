package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// EndGameHandler handles WebSocket messages to end a game.
type EndGameHandler struct {
	action *connaction.EndGameAction
	hub    *core.Hub
	logger *slog.Logger
}

// NewEndGameHandler creates a new EndGameHandler.
func NewEndGameHandler(action *connaction.EndGameAction, hub *core.Hub) *EndGameHandler {
	return &EndGameHandler{
		action: action,
		hub:    hub,
		logger: logger.Get(),
	}
}

// HandleMessage processes an end-game WebSocket message.
func (h *EndGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing end game request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "not connected to game")
		return
	}

	gameID := connection.GameID()

	err := h.action.Execute(ctx, gameID, connection.PlayerID())
	if err != nil {
		log.Error("Failed to execute end game action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Game ended")

	gameEndedMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypeGameEnded,
		GameID:  gameID,
		Payload: map[string]any{"reason": "The host ended the game"},
	}

	h.hub.DisconnectGame(gameID, gameEndedMessage)
}
