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

// Broadcaster defines the interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// TradeHandler handles colony trade requests
type TradeHandler struct {
	action      *colonyaction.TradeAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewTradeHandler creates a new colony trade handler
func NewTradeHandler(action *colonyaction.TradeAction, broadcaster Broadcaster) *TradeHandler {
	return &TradeHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *TradeHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing colony trade request")

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

	var payload dto.ColonyTradeRequest
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	if payload.ColonyID == "" {
		log.Error("Missing colony ID in payload")
		h.sendError(connection, "Colony ID is required")
		return
	}

	if payload.TrackSteps == nil {
		h.sendError(connection, "trackSteps is required")
		return
	}
	paymentType := colonyaction.TradePaymentType(payload.PaymentType)
	if paymentType == "" {
		paymentType = colonyaction.TradePaymentEnergy
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

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, payload.ColonyID, paymentType, *payload.TrackSteps, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute colony trade action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Colony traded", slog.String("colony_id", payload.ColonyID))

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":   "colony-trade",
			"colonyId": payload.ColonyID,
			"success":  true,
		},
	}
}

func (h *TradeHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
