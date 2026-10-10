package confirmation

import (
	"context"
	"encoding/json"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmFreeTradeHandler handles confirm free trade requests
type ConfirmFreeTradeHandler struct {
	action      *confirmaction.ConfirmFreeTradeAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmFreeTradeHandler creates a new confirm free trade handler
func NewConfirmFreeTradeHandler(action *confirmaction.ConfirmFreeTradeAction, broadcaster Broadcaster) *ConfirmFreeTradeHandler {
	return &ConfirmFreeTradeHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmFreeTradeHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm free trade request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		h.sendError(connection, "Invalid payload format")
		return
	}
	var payload dto.FreeTradeRequest
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		h.sendError(connection, "Invalid payload format")
		return
	}
	colonyID := payload.ColonyID
	if colonyID == "" {
		h.sendError(connection, "Missing colonyId in payload")
		return
	}
	if payload.TrackSteps == nil {
		h.sendError(connection, "trackSteps is required")
		return
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, colonyID, *payload.TrackSteps)
	if err != nil {
		log.Error("Failed to execute confirm free trade action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Free trade confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-free-trade",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmFreeTradeHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
