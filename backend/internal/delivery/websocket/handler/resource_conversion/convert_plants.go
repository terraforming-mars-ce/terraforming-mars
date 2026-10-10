package resource_conversion

import (
	"context"
	"encoding/json"
	"log/slog"

	resconvaction "openmars/internal/action/resource_conversion"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConvertPlantsHandler handles convert plants to greenery requests
type ConvertPlantsHandler struct {
	action      *resconvaction.ConvertPlantsToGreeneryAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConvertPlantsHandler creates a new convert plants handler
func NewConvertPlantsHandler(action *resconvaction.ConvertPlantsToGreeneryAction, broadcaster Broadcaster) *ConvertPlantsHandler {
	return &ConvertPlantsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConvertPlantsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing convert plants request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	var req dto.ActionConvertPlantsToGreeneryRequest
	if message.Payload != nil {
		payloadBytes, _ := json.Marshal(message.Payload)
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			connection.SendError(message.Type, "Invalid payment")
			return
		}
	}

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), dto.ToPayment(req.Payment))
	if err != nil {
		log.Error("Failed to execute convert plants action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Plant conversion completed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
