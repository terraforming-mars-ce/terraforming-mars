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

// ConfirmResourceRemovalHandler handles confirm resource removal requests
type ConfirmResourceRemovalHandler struct {
	action      *confirmaction.ConfirmResourceRemovalAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmResourceRemovalHandler creates a new confirm resource removal handler
func NewConfirmResourceRemovalHandler(action *confirmaction.ConfirmResourceRemovalAction, broadcaster Broadcaster) *ConfirmResourceRemovalHandler {
	return &ConfirmResourceRemovalHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmResourceRemovalHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm resource removal request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	selectionID, _ := payloadMap["selectionId"].(string)
	fail := func(message string) {
		connection.Send <- dto.WebSocketMessage{Type: dto.MessageTypeError, Payload: map[string]interface{}{"error": message, "selectionId": selectionID}}
	}
	var payload struct {
		SelectionID    string `json:"selectionId"`
		TargetPlayerID string `json:"targetPlayerId"`
		Amount         *int   `json:"amount"`
	}
	encoded, err := json.Marshal(payloadMap)
	if err != nil {
		fail("Invalid payload")
		return
	}
	if err := json.Unmarshal(encoded, &payload); err != nil || payload.Amount == nil || payload.SelectionID == "" {
		fail("Invalid selectionId or integer amount")
		return
	}
	if err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, payload.SelectionID, payload.TargetPlayerID, *payload.Amount); err != nil {
		fail(err.Error())
		return
	}

	log.Debug("Resource removal confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-resource-removal",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmResourceRemovalHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
