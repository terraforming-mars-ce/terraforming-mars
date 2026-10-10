package tile

import (
	"context"
	"log/slog"

	tileaction "openmars/internal/action/tile"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// SelectTileHandler handles tile selection requests
type SelectTileHandler struct {
	action      *tileaction.SelectTileAction
	broadcaster Broadcaster
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing tile selection request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	selectedHex, ok := payload["hex"].(string)
	if !ok || selectedHex == "" {
		log.Error("Missing or invalid hex")
		connection.SendError(message.Type, "Missing hex position")
		return
	}

	log.Debug("Hex position extracted", slog.String("hex", selectedHex))

	_, err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), selectedHex)
	if err != nil {
		log.Error("Failed to execute select tile action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Tile selection completed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
