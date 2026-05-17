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

// UpdateGameSettingsHandler handles incoming update-game-settings messages.
type UpdateGameSettingsHandler struct {
	action      *gameaction.UpdateGameSettingsAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewUpdateGameSettingsHandler creates a new update game settings handler.
func NewUpdateGameSettingsHandler(action *gameaction.UpdateGameSettingsAction, broadcaster Broadcaster) *UpdateGameSettingsHandler {
	return &UpdateGameSettingsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface.
func (h *UpdateGameSettingsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	gameID := connection.GameID
	playerID := connection.PlayerID
	if gameID == "" || playerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	// Round-trip through JSON so we get pointer fields populated only for keys
	// that the client actually sent.
	raw, err := json.Marshal(message.Payload)
	if err != nil {
		h.sendError(connection, "Invalid payload")
		return
	}
	var patch dto.UpdateGameSettingsRequest
	if err := json.Unmarshal(raw, &patch); err != nil {
		h.sendError(connection, "Invalid payload")
		return
	}

	if err := h.action.Execute(ctx, gameID, playerID, &patch); err != nil {
		log.Debug("Failed to update game settings", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Game settings updated and broadcast")
}

func (h *UpdateGameSettingsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	}
}
