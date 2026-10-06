package game

import (
	"context"
	"encoding/json"
	"log/slog"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
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
	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "not connected to game")
		return
	}
	raw, err := json.Marshal(message.Payload)
	if err != nil {
		h.sendError(connection, "invalid payload")
		return
	}
	var payload dto.BotInspectPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		h.sendError(connection, "invalid payload")
		return
	}
	if err := h.bots.InspectBot(ctx, connection.GameID, connection.PlayerID, connection.ID, payload.PlayerID, connection.SendMessage, connection.Done); err != nil {
		h.logger.Debug("Bot inspection rejected", slog.String("game_id", connection.GameID), slog.Any("error", err))
		h.sendError(connection, err.Error())
	}
}

func (h *BotInspectHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type:    dto.MessageTypeError,
		Payload: map[string]interface{}{"error": errorMessage},
	})
}
