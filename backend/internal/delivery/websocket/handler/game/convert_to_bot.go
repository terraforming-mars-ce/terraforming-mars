package game

import (
	"context"
	"log/slog"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

type ConvertToBotHandler struct {
	action      *gameaction.ConvertToBotAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *slog.Logger
}

func NewConvertToBotHandler(action *gameaction.ConvertToBotAction, broadcaster Broadcaster, hub *core.Hub) *ConvertToBotHandler {
	return &ConvertToBotHandler{
		action:      action,
		broadcaster: broadcaster,
		hub:         hub,
		logger:      logger.Get(),
	}
}

func (h *ConvertToBotHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)
	log.Debug("Processing convert to bot request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		connection.SendError(message.Type, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]any)
	if !ok {
		connection.SendError(message.Type, "invalid payload format")
		return
	}

	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)
	if targetPlayerID == "" {
		connection.SendError(message.Type, "targetPlayerId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), targetPlayerID)
	if err != nil {
		log.Error("Failed to convert player to bot", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Player converted to bot")

	h.hub.DisconnectPlayer(connection.GameID(), targetPlayerID, dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerKicked,
		GameID:  connection.GameID(),
		Payload: map[string]any{"reason": "You were replaced by a bot"},
	})

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
}
