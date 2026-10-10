package connection

import (
	"context"
	"log/slog"
	"time"

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

	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)
	if targetPlayerID == "" {
		log.Error("Missing targetPlayerId in payload")
		h.sendError(connection, "targetPlayerId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, targetPlayerID)
	if err != nil {
		log.Error("Failed to execute kick player action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Player kicked")

	// Send player-kicked message to the kicked player before closing their connection
	kickedConnection := h.hub.GetManager().GetConnectionByPlayerID(connection.GameID, targetPlayerID)
	if kickedConnection != nil {
		kickedMessage := dto.WebSocketMessage{
			Type:    dto.MessageTypePlayerKicked,
			GameID:  connection.GameID,
			Payload: map[string]any{"reason": "You were kicked from the game"},
		}
		kickedConnection.SendMessage(kickedMessage)
		log.Debug("Sent player-kicked message to kicked player", slog.String("target_player_id", targetPlayerID))

		// Close the kicked player's connection after a short delay to ensure the message is sent
		go func() {
			time.Sleep(100 * time.Millisecond)
			kickedConnection.Close()
			log.Debug("Closed kicked player's connection", slog.String("target_player_id", targetPlayerID))
		}()
	}

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")
}

func (h *KickPlayerHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	}
}
