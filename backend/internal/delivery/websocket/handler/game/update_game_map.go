package game

import (
	"context"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// UpdateGameMapHandler handles update game map requests
type UpdateGameMapHandler struct {
	updateGameMapAction *gameaction.UpdateGameMapAction
	broadcaster         Broadcaster
	logger              *zap.Logger
}

// NewUpdateGameMapHandler creates a new handler
func NewUpdateGameMapHandler(action *gameaction.UpdateGameMapAction, broadcaster Broadcaster) *UpdateGameMapHandler {
	return &UpdateGameMapHandler{
		updateGameMapAction: action,
		broadcaster:         broadcaster,
		logger:              logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *UpdateGameMapHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		h.sendError(connection, "Invalid payload")
		return
	}

	mapID, ok := payloadMap["mapId"].(string)
	if !ok || mapID == "" {
		h.sendError(connection, "Missing mapId")
		return
	}

	gameID := connection.GameID
	playerID := connection.PlayerID

	if gameID == "" || playerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.updateGameMapAction.Execute(ctx, gameID, playerID, mapID)
	if err != nil {
		log.Debug("Failed to update game map", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Game map updated and broadcast")
}

func (h *UpdateGameMapHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
