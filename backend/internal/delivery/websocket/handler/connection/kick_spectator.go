package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// KickSpectatorHandler handles kicking a spectator from a game.
type KickSpectatorHandler struct {
	action      *connaction.KickSpectatorAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *slog.Logger
}

// NewKickSpectatorHandler creates a new kick spectator handler.
func NewKickSpectatorHandler(action *connaction.KickSpectatorAction, broadcaster Broadcaster, hub *core.Hub) *KickSpectatorHandler {
	return &KickSpectatorHandler{
		action:      action,
		broadcaster: broadcaster,
		hub:         hub,
		logger:      logger.Get(),
	}
}

// HandleMessage processes a kick-spectator message.
func (h *KickSpectatorHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing kick spectator request")

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

	targetSpectatorID, _ := payloadMap["targetSpectatorId"].(string)
	if targetSpectatorID == "" {
		log.Error("Missing targetSpectatorId in payload")
		connection.SendError(message.Type, "targetSpectatorId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), targetSpectatorID)
	if err != nil {
		log.Error("Failed to kick spectator", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Spectator kicked")

	h.hub.DisconnectSpectator(connection.GameID(), targetSpectatorID, dto.WebSocketMessage{
		Type:    dto.MessageTypeSpectatorKicked,
		GameID:  connection.GameID(),
		Payload: map[string]any{"reason": "You were kicked from the game"},
	})

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
}
