package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// PlayerTakeoverHandler handles player takeover requests
type PlayerTakeoverHandler struct {
	action      *connaction.PlayerTakeoverAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewPlayerTakeoverHandler creates a new player takeover handler
func NewPlayerTakeoverHandler(action *connaction.PlayerTakeoverAction, broadcaster Broadcaster) *PlayerTakeoverHandler {
	return &PlayerTakeoverHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *PlayerTakeoverHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing player takeover request")

	payloadMap, ok := message.Payload.(map[string]any)
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		connection.SendError(message.Type, "Missing gameId")
		return
	}

	if targetPlayerID == "" {
		log.Error("Missing targetPlayerId")
		connection.SendError(message.Type, "Missing targetPlayerId")
		return
	}

	log.Debug("Parsed takeover request",
		slog.String("game_id", gameID),
		slog.String("target_player_id", targetPlayerID))

	result, err := h.action.Execute(ctx, gameID, targetPlayerID)
	if err != nil {
		log.Error("Failed to execute player takeover action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	connection.BindPlayer(gameID, result.PlayerID)
	h.broadcaster.BroadcastGameState(gameID, nil)
	h.broadcaster.SendInitialLogs(gameID, result.PlayerID)
	connection.Send(dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerConnected,
		GameID:  gameID,
		Payload: dto.PlayerConnectedPayload{PlayerID: result.PlayerID, PlayerName: result.PlayerName},
	})
}
