package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

type KickPlayerHandler struct {
	action      *connaction.KickPlayerAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *slog.Logger
}

func NewKickPlayerHandler(action *connaction.KickPlayerAction, broadcaster Broadcaster, hub *core.Hub) *KickPlayerHandler {
	return &KickPlayerHandler{
		action:      action,
		broadcaster: broadcaster,
		hub:         hub,
		logger:      logger.Get(),
	}
}

func (h *KickPlayerHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing kick player request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]any)
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "invalid payload format")
		return
	}

	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)
	if targetPlayerID == "" {
		log.Error("Missing targetPlayerId in payload")
		connection.SendError(message.Type, "targetPlayerId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), targetPlayerID)
	if err != nil {
		log.Error("Failed to execute kick player action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Player kicked")

	h.hub.DisconnectPlayer(connection.GameID(), targetPlayerID, dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerKicked,
		GameID:  connection.GameID(),
		Payload: map[string]any{"reason": "You were kicked from the game"},
	})

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")
}
