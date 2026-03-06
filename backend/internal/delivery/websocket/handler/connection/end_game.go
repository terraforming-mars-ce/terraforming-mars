package connection

import (
	"context"
	"time"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// EndGameHandler handles WebSocket messages to end a game.
type EndGameHandler struct {
	action *connaction.EndGameAction
	hub    *core.Hub
	logger *zap.Logger
}

// NewEndGameHandler creates a new EndGameHandler.
func NewEndGameHandler(action *connaction.EndGameAction, hub *core.Hub) *EndGameHandler {
	return &EndGameHandler{
		action: action,
		hub:    hub,
		logger: logger.Get(),
	}
}

// HandleMessage processes an end-game WebSocket message.
func (h *EndGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing end game request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "not connected to game")
		return
	}

	gameID := connection.GameID

	gameConnections := h.hub.GetManager().GetGameConnections(gameID)

	err := h.action.Execute(ctx, gameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute end game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Game ended")

	gameEndedMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypeGameEnded,
		GameID:  gameID,
		Payload: map[string]any{"reason": "The host ended the game"},
	}

	for conn := range gameConnections {
		conn.SendMessage(gameEndedMessage)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		for conn := range gameConnections {
			conn.Close()
		}
		log.Debug("Closed all connections for ended game", zap.String("game_id", gameID))
	}()
}

func (h *EndGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	}
}
