package utils

import (
	"context"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// StandardProjectHandler provides common functionality for standard project WebSocket handlers
type StandardProjectHandler struct {
	parser       *MessageParser
	errorHandler *ErrorHandler
	logger       *zap.Logger
}

// NewStandardProjectHandler creates a new StandardProjectHandler base
func NewStandardProjectHandler(parser *MessageParser) *StandardProjectHandler {
	return &StandardProjectHandler{
		parser:       parser,
		errorHandler: NewErrorHandler(),
		logger:       logger.Get(),
	}
}

// HandleStandardProject executes a standard project with common validation and error handling
func (h *StandardProjectHandler) HandleStandardProject(
	ctx context.Context,
	connection *core.Connection,
	projectName string,
	emoji string,
	projectFunc func(ctx context.Context, gameID, playerID string) error,
) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn(projectName+" action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, ErrMustConnectFirst)
		return
	}

	h.logger.Debug(emoji+" Processing "+projectName+" action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	if err := projectFunc(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to "+projectName,
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Debug(projectName+" completed, tile queued",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
