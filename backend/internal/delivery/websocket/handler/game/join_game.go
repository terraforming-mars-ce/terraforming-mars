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
		h.sendError(connection, "Invalid payload format")
		return
	}

	gameID, _ := payloadMap["gameId"].(string)
	playerName, _ := payloadMap["playerName"].(string)
	playerID, _ := payloadMap["playerId"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		h.sendError(connection, "Missing gameId")
		return
	}

	if playerName == "" {
		log.Error("Missing playerName")
		h.sendError(connection, "Missing playerName")
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
		log.Error("Failed to execute join game action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	connection.SetPlayer(result.PlayerID, gameID)
	log.Debug("Game joined",
		slog.String("player_id", result.PlayerID))

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Broadcasted game state to all players")

	h.broadcaster.SendInitialLogs(gameID, playerID)
	log.Debug("Sent initial logs to player")

	response := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnected,
		GameID: gameID,
		Payload: map[string]interface{}{
			"playerID":   result.PlayerID,
			"playerName": playerName,
			"success":    true,
		},
	}

	connection.Send <- response
	log.Debug("Sent player connected confirmation")
}

// sendError sends an error message to the client
func (h *JoinGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
