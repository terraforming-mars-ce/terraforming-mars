package game

import (
	"context"
	"encoding/json"
	"log/slog"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// BotInspector streams a bot's trace to an admin connection.
type BotInspector interface {
	InspectBot(ctx context.Context, gameID, requesterID, connectionID, playerID string, send func(dto.WebSocketMessage), done <-chan struct{}) error
}

// BotInspectHandler subscribes the host's admin tools to a bot's trace.
type BotInspectHandler struct {
	bots   BotInspector
	logger *slog.Logger
}

// NewBotInspectHandler creates a bot inspect handler.
func NewBotInspectHandler(bots BotInspector) *BotInspectHandler {
	return &BotInspectHandler{bots: bots, logger: logger.Get()}
}

// HandleMessage processes a bot-inspect request.
func (h *BotInspectHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	if connection.GameID() == "" || connection.PlayerID() == "" {
		connection.SendError(message.Type, "not connected to game")
		return
	}
	raw, err := json.Marshal(message.Payload)
	if err != nil {
		connection.SendError(message.Type, "invalid payload")
		return
	}
	var payload dto.BotInspectPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		connection.SendError(message.Type, "invalid payload")
		return
	}
	if err := h.bots.InspectBot(ctx, connection.GameID(), connection.PlayerID(), connection.ID, payload.PlayerID, connection.Send, connection.Done()); err != nil {
		h.logger.Debug("Bot inspection rejected", slog.String("game_id", connection.GameID()), slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
	}
}
