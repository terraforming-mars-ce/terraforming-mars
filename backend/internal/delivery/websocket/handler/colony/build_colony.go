package colony

import (
	"context"
	"encoding/json"
	"log/slog"
	"openmars/internal/game/shared"

	colonyaction "openmars/internal/action/colony"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// BuildColonyPayload represents the expected payload for building a colony
type BuildColonyPayload struct {
	ColonyID string `json:"colonyId"`
}

// BuildColonyHandler handles build colony requests
type BuildColonyHandler struct {
	action      *colonyaction.BuildColonyAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewBuildColonyHandler creates a new build colony handler
func NewBuildColonyHandler(action *colonyaction.BuildColonyAction, broadcaster Broadcaster) *BuildColonyHandler {
	return &BuildColonyHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *BuildColonyHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing build colony request")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", slog.Any("error", err))
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	var payload BuildColonyPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	if payload.ColonyID == "" {
		log.Error("Missing colony ID in payload")
		connection.SendError(message.Type, "Colony ID is required")
		return
	}

	var paymentEnvelope struct {
		Payment shared.Payment `json:"payment"`
	}
	paymentBytes, paymentErr := json.Marshal(message.Payload)
	if paymentErr != nil {
		connection.SendError(message.Type, "Invalid payment")
		return
	}
	if paymentErr = json.Unmarshal(paymentBytes, &paymentEnvelope); paymentErr != nil {
		connection.SendError(message.Type, "Invalid payment")
		return
	}

	err = h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), payload.ColonyID, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute build colony action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Colony built", slog.String("colony_id", payload.ColonyID))

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}
