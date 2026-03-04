package connection

import (
	"context"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SpectatorDisconnectedHandler handles spectator disconnect events.
type SpectatorDisconnectedHandler struct {
	action      *connaction.SpectatorDisconnectedAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSpectatorDisconnectedHandler creates a new spectator disconnected handler.
func NewSpectatorDisconnectedHandler(action *connaction.SpectatorDisconnectedAction, broadcaster Broadcaster) *SpectatorDisconnectedHandler {
	return &SpectatorDisconnectedHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage processes a spectator-disconnected message.
func (h *SpectatorDisconnectedHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		payload, ok2 := message.Payload.(dto.SpectatorDisconnectedPayload)
		if !ok2 {
			log.Error("Invalid payload format for spectator disconnect")
			return
		}
		payloadMap = map[string]interface{}{
			"spectatorId": payload.SpectatorID,
			"gameId":      payload.GameID,
		}
	}

	spectatorID, _ := payloadMap["spectatorId"].(string)
	gameID, _ := payloadMap["gameId"].(string)

	if spectatorID == "" || gameID == "" {
		log.Debug("Missing spectator/game ID in disconnect payload")
		return
	}

	log = log.With(
		zap.String("spectator_id", spectatorID),
		zap.String("game_id", gameID),
	)

	if err := h.action.Execute(ctx, gameID, spectatorID); err != nil {
		log.Error("Failed to handle spectator disconnect", zap.Error(err))
		return
	}

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("📡 Broadcasted updated state after spectator disconnect")
}
