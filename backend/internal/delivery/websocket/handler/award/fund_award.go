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

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	var payload FundAwardPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	if payload.AwardType == "" {
		log.Error("Missing award type in payload")
		h.sendError(connection, "Award type is required")
		return
	}

	var paymentEnvelope struct {
		Payment shared.Payment `json:"payment"`
	}
	paymentBytes, paymentErr := json.Marshal(message.Payload)
	if paymentErr != nil {
		h.sendError(connection, "Invalid payment")
		return
	}
	if paymentErr = json.Unmarshal(paymentBytes, &paymentEnvelope); paymentErr != nil {
		h.sendError(connection, "Invalid payment")
		return
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, payload.AwardType, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute fund award action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Fund award completed",
		slog.String("award_type", payload.AwardType))

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":    "fund-award",
			"awardType": payload.AwardType,
			"success":   true,
		},
	}

	connection.Send <- response
}

// sendError sends an error message to the client
func (h *FundAwardHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
