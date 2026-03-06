package connection

import (
	"context"
	"time"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// KickSpectatorHandler handles kicking a spectator from a game.
type KickSpectatorHandler struct {
	action      *connaction.KickSpectatorAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *zap.Logger
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing kick spectator request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]any)
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "invalid payload format")
		return
	}

	targetSpectatorID, _ := payloadMap["targetSpectatorId"].(string)
	if targetSpectatorID == "" {
		log.Error("Missing targetSpectatorId in payload")
		h.sendError(connection, "targetSpectatorId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, targetSpectatorID)
	if err != nil {
		log.Error("Failed to kick spectator", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Spectator kicked")

	kickedConn := h.hub.GetManager().GetConnectionBySpectatorID(connection.GameID, targetSpectatorID)
	if kickedConn != nil {
		kickedMessage := dto.WebSocketMessage{
			Type:    dto.MessageTypeSpectatorKicked,
			GameID:  connection.GameID,
			Payload: map[string]any{"reason": "You were kicked from the game"},
		}
		kickedConn.SendMessage(kickedMessage)

		go func() {
			time.Sleep(100 * time.Millisecond)
			kickedConn.Close()
		}()
	}

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
}

func (h *KickSpectatorHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	})
}
