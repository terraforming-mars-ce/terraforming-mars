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
		h.sendError(connection, "Invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		h.sendError(connection, "Missing gameId")
		return
	}

	if targetPlayerID == "" {
		log.Error("Missing targetPlayerId")
		h.sendError(connection, "Missing targetPlayerId")
		return
	}

	log.Debug("Parsed takeover request",
		slog.String("game_id", gameID),
		slog.String("target_player_id", targetPlayerID))

	result, err := h.action.Execute(ctx, gameID, targetPlayerID)
	if err != nil {
		log.Error("Failed to execute player takeover action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	connection.SetPlayer(targetPlayerID, gameID)

	log.Debug("Player takeover completed",
		slog.String("player_id", result.PlayerID),
		slog.String("player_name", result.PlayerName))

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnected,
		GameID: gameID,
		Payload: map[string]any{
			"playerID":   result.PlayerID,
			"playerName": result.PlayerName,
			"success":    true,
		},
	}

	connection.Send <- response
	log.Debug("Sent player takeover confirmation")
}

// sendError sends an error message to the client
func (h *PlayerTakeoverHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	}
}
