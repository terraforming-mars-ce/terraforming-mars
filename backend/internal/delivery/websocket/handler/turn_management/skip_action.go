package turn_management

import (
	"context"
	"log/slog"

	turnaction "openmars/internal/action/turn_management"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// SkipActionHandler handles skip action requests
type SkipActionHandler struct {
	action      *turnaction.SkipActionAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewSkipActionHandler creates a new skip action handler
func NewSkipActionHandler(action *turnaction.SkipActionAction, broadcaster Broadcaster) *SkipActionHandler {
	return &SkipActionHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SkipActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing skip action request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute skip action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Skip action completed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "skip-action",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SkipActionHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
