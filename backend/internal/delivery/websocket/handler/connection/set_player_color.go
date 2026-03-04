package connection

import (
	"context"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SetPlayerColorHandler handles set-player-color messages.
type SetPlayerColorHandler struct {
	action      *connaction.SetPlayerColorAction
	broadcaster Broadcaster
	logger      *zap.Logger
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	gameID := connection.GameID
	playerID := connection.PlayerID
	if gameID == "" || playerID == "" {
		h.sendError(connection, "not connected to game as a player")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "invalid payload format")
		return
	}

	color, _ := payloadMap["color"].(string)
	if color == "" {
		h.sendError(connection, "missing color")
		return
	}

	targetPlayerID := playerID
	if tid, ok := payloadMap["targetPlayerId"].(string); ok && tid != "" {
		targetPlayerID = tid
	}

	if err := h.action.Execute(ctx, gameID, playerID, targetPlayerID, color); err != nil {
		log.Error("Failed to set player color", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(gameID, nil)
}

func (h *SetPlayerColorHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	})
}
