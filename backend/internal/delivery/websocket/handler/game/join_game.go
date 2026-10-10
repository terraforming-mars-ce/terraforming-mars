package game

import (
	"context"
	"log/slog"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"

	"github.com/google/uuid"
)

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	SendInitialLogs(gameID string, playerID string)
}

// JoinGameHandler handles join game requests.
type JoinGameHandler struct {
	joinGameAction *gameaction.JoinGameAction
	broadcaster    Broadcaster
	logger         *slog.Logger
}

// NewJoinGameHandler creates a new join game handler
func NewJoinGameHandler(
	joinGameAction *gameaction.JoinGameAction,
	broadcaster Broadcaster,
) *JoinGameHandler {
	return &JoinGameHandler{
		joinGameAction: joinGameAction,
		broadcaster:    broadcaster,
		logger:         logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *JoinGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing join game request")

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	playerName, _ := payloadMap["playerName"].(string)
	playerID, _ := payloadMap["playerId"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		connection.SendError(message.Type, "Missing gameId")
		return
	}

	if playerName == "" {
		log.Error("Missing playerName")
		connection.SendError(message.Type, "Missing playerName")
		return
	}

	if playerID == "" {
		playerID = uuid.New().String()
		log.Debug("Generated new playerID for session",
			slog.String("player_id", playerID))
	}

	log.Debug("Parsed join game request",
		slog.String("game_id", gameID),
		slog.String("player_name", playerName),
		slog.String("player_id", playerID))

	result, err := h.joinGameAction.Execute(ctx, gameID, playerName, playerID)
	if err != nil {
		log.Warn("Failed to join game", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	connection.BindPlayer(gameID, result.PlayerID)
	h.broadcaster.BroadcastGameState(gameID, nil)
	h.broadcaster.SendInitialLogs(gameID, result.PlayerID)
	connection.Send(dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerConnected,
		GameID:  gameID,
		Payload: dto.PlayerConnectedPayload{PlayerID: result.PlayerID, PlayerName: playerName},
	})
}
