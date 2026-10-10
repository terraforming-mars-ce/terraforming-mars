package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// SetPlayerColorHandler handles set-player-color messages.
type SetPlayerColorHandler struct {
	action      *connaction.SetPlayerColorAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewSetPlayerColorHandler creates a new set player color handler.
func NewSetPlayerColorHandler(action *connaction.SetPlayerColorAction, broadcaster Broadcaster) *SetPlayerColorHandler {
	return &SetPlayerColorHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage processes a set-player-color message.
func (h *SetPlayerColorHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	gameID := connection.GameID()
	playerID := connection.PlayerID()
	if gameID == "" || playerID == "" {
		connection.SendError(message.Type, "not connected to game as a player")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "invalid payload format")
		return
	}

	color, _ := payloadMap["color"].(string)
	if color == "" {
		connection.SendError(message.Type, "missing color")
		return
	}

	targetPlayerID := playerID
	if tid, ok := payloadMap["targetPlayerId"].(string); ok && tid != "" {
		targetPlayerID = tid
	}

	if err := h.action.Execute(ctx, gameID, playerID, targetPlayerID, color); err != nil {
		log.Error("Failed to set player color", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(gameID, nil)
}
