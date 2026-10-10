package game

import (
	"context"
	"encoding/json"
	"log/slog"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// BotRetrier lets the host give a failed bot another chance with the model.
type BotRetrier interface {
	RetryBot(ctx context.Context, gameID, requesterID, playerID string) error
}

// BotRetryHandler handles the host's request to retry a failed bot.
type BotRetryHandler struct {
	bots   BotRetrier
	logger *slog.Logger
}

// NewBotRetryHandler creates a bot retry handler.
func NewBotRetryHandler(bots BotRetrier) *BotRetryHandler {
	return &BotRetryHandler{bots: bots, logger: logger.Get()}
}

// HandleMessage processes a bot-retry request.
func (h *BotRetryHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "not connected to game")
		return
	}
	raw, err := json.Marshal(message.Payload)
	if err != nil {
		h.sendError(connection, "invalid payload")
		return
	}
	var payload dto.BotRetryPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.PlayerID == "" {
		h.sendError(connection, "invalid payload")
		return
	}
	if err := h.bots.RetryBot(ctx, connection.GameID, connection.PlayerID, payload.PlayerID); err != nil {
		h.logger.Warn("Bot retry rejected", slog.String("game_id", connection.GameID), slog.Any("error", err))
		h.sendError(connection, err.Error())
	}
}

func (h *BotRetryHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type:    dto.MessageTypeError,
		Payload: map[string]interface{}{"error": errorMessage},
	})
}
