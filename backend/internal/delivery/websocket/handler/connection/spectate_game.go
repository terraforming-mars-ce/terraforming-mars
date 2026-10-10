package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"

	"github.com/google/uuid"
)

// SpectateGameHandler handles spectator connection requests.
type SpectateGameHandler struct {
	action      *connaction.SpectateGameAction
	broadcaster Broadcaster
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing spectate game request")

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	spectatorName, _ := payloadMap["spectatorName"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		connection.SendError(message.Type, "missing gameId")
		return
	}

	if spectatorName == "" {
		log.Error("Missing spectatorName")
		connection.SendError(message.Type, "missing spectatorName")
		return
	}

	if current := connection.Identity(); current.GameID == gameID && current.IsSpectator() {
		h.broadcaster.SendInitialLogsToSpectator(gameID, current.SpectatorID)
		h.broadcaster.BroadcastGameState(gameID, nil)
		connection.Send(spectatorConnected(gameID, current.SpectatorID))
		return
	}

	result, err := h.action.Execute(ctx, gameID, spectatorName, uuid.New().String())
	if err != nil {
		log.Warn("Failed to spectate game", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	connection.BindSpectator(gameID, result.SpectatorID)
	h.broadcaster.BroadcastGameState(gameID, nil)
	h.broadcaster.SendInitialLogsToSpectator(gameID, result.SpectatorID)
	connection.Send(spectatorConnected(gameID, result.SpectatorID))
}

func spectatorConnected(gameID, spectatorID string) dto.WebSocketMessage {
	return dto.WebSocketMessage{
		Type:    dto.MessageTypeSpectatorConnected,
		GameID:  gameID,
		Payload: dto.SpectatorConnectedPayload{SpectatorID: spectatorID},
	}
}
