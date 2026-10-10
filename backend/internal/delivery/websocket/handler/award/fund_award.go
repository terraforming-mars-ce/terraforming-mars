package award

import (
	"context"
	"encoding/json"
	"log/slog"
	"openmars/internal/game/shared"

	awardaction "openmars/internal/action/award"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// Broadcaster defines the interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// FundAwardHandler handles fund award requests
type FundAwardHandler struct {
	action      *awardaction.FundAwardAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewFundAwardHandler creates a new fund award handler
func NewFundAwardHandler(action *awardaction.FundAwardAction, broadcaster Broadcaster) *FundAwardHandler {
	return &FundAwardHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// FundAwardPayload represents the expected payload for funding an award
type FundAwardPayload struct {
	AwardType string `json:"awardType"`
}

// HandleMessage implements the MessageHandler interface
func (h *FundAwardHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing fund award request")

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

	var payload FundAwardPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	if payload.AwardType == "" {
		log.Error("Missing award type in payload")
		connection.SendError(message.Type, "Award type is required")
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

	err = h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), payload.AwardType, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute fund award action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Fund award completed",
		slog.String("award_type", payload.AwardType))

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)
	log.Debug("Broadcasted game state to all players")

}
