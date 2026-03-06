package standard_project

import (
	"context"

	stdprojaction "terraforming-mars-backend/internal/action/standard_project"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// LaunchAsteroidHandler handles launch asteroid standard project requests
type LaunchAsteroidHandler struct {
	action      *stdprojaction.LaunchAsteroidAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewLaunchAsteroidHandler creates a new launch asteroid handler
func NewLaunchAsteroidHandler(action *stdprojaction.LaunchAsteroidAction, broadcaster Broadcaster) *LaunchAsteroidHandler {
	return &LaunchAsteroidHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *LaunchAsteroidHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing launch asteroid request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute launch asteroid action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Asteroid launched")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "launch-asteroid",
			"success": true,
		},
	}

	connection.Send <- response
}

// sendError sends an error message to the client
func (h *LaunchAsteroidHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
