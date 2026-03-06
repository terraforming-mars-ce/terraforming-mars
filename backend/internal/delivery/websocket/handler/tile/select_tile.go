package tile

import (
	"context"

	tileaction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SelectTileHandler handles tile selection requests
type SelectTileHandler struct {
	action      *tileaction.SelectTileAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewSelectTileHandler creates a new select tile handler
func NewSelectTileHandler(action *tileaction.SelectTileAction, broadcaster Broadcaster) *SelectTileHandler {
	return &SelectTileHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectTileHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing tile selection request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	selectedHex, ok := payload["hex"].(string)
	if !ok || selectedHex == "" {
		log.Error("Missing or invalid hex")
		h.sendError(connection, "Missing hex position")
		return
	}

	log.Debug("Hex position extracted", zap.String("hex", selectedHex))

	_, err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, selectedHex)
	if err != nil {
		log.Error("Failed to execute select tile action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Tile selection completed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-tile",
			"success": true,
			"hex":     selectedHex,
		},
	}

	connection.Send <- response
}

func (h *SelectTileHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
