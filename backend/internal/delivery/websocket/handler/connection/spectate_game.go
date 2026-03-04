package connection

import (
	"context"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SpectateGameHandler handles spectator connection requests.
type SpectateGameHandler struct {
	action      *connaction.SpectateGameAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSpectateGameHandler creates a new spectate game handler.
func NewSpectateGameHandler(action *connaction.SpectateGameAction, broadcaster Broadcaster) *SpectateGameHandler {
	return &SpectateGameHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage processes a spectator-connect message.
func (h *SpectateGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("👁️ Processing spectate game request")

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	spectatorName, _ := payloadMap["spectatorName"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		h.sendError(connection, "missing gameId")
		return
	}

	if spectatorName == "" {
		log.Error("Missing spectatorName")
		h.sendError(connection, "missing spectatorName")
		return
	}

	spectatorID := uuid.New().String()
	connection.SetSpectator(spectatorID, gameID)

	result, err := h.action.Execute(ctx, gameID, spectatorName, spectatorID)
	if err != nil {
		log.Error("Failed to execute spectate game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Spectator joined successfully", zap.String("spectator_id", result.SpectatorID))

	h.broadcaster.BroadcastGameState(gameID, nil)

	h.broadcaster.SendInitialLogsToSpectator(gameID, spectatorID)

	response := dto.WebSocketMessage{
		Type:   dto.MessageTypeSpectatorConnected,
		GameID: gameID,
		Payload: map[string]interface{}{
			"spectatorId": result.SpectatorID,
			"success":     true,
		},
	}
	connection.SendMessage(response)
}

func (h *SpectateGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	})
}
