package game

import (
	"context"
	"encoding/json"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// CreateGameHandler handles create game requests.
type CreateGameHandler struct {
	createGameAction *gameaction.CreateGameAction
	broadcaster      Broadcaster
	logger           *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	SendInitialLogs(gameID string, playerID string)
}

// NewCreateGameHandler creates a new create game handler.
func NewCreateGameHandler(createGameAction *gameaction.CreateGameAction, broadcaster Broadcaster) *CreateGameHandler {
	return &CreateGameHandler{
		createGameAction: createGameAction,
		broadcaster:      broadcaster,
		logger:           logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *CreateGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing create game request")

	var req dto.CreateGameRequest
	payload, err := json.Marshal(message.Payload)
	if err != nil {
		h.sendError(connection, "Invalid game settings")
		return
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendError(connection, "Invalid game settings")
		return
	}
	game, err := h.createGameAction.ExecuteRequest(ctx, req)
	if err != nil {
		log.Error("Failed to execute create game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Game created",
		zap.String("game_id", game.ID()))

	h.broadcaster.BroadcastGameState(game.ID(), nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: map[string]interface{}{
			"gameId":  game.ID(),
			"success": true,
			"message": "Game created. Join with playerConnect.",
		},
	}

	connection.Send <- response

	log.Debug("Sent game created response to client")
}

// sendError sends an error message to the client
func (h *CreateGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
