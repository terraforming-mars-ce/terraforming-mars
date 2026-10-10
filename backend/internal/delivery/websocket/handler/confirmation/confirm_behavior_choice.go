package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmBehaviorChoiceHandler handles confirm behavior choice requests
type ConfirmBehaviorChoiceHandler struct {
	action      *confirmaction.ConfirmBehaviorChoiceAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmBehaviorChoiceHandler creates a new confirm behavior choice handler
func NewConfirmBehaviorChoiceHandler(action *confirmaction.ConfirmBehaviorChoiceAction, broadcaster Broadcaster) *ConfirmBehaviorChoiceHandler {
	return &ConfirmBehaviorChoiceHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmBehaviorChoiceHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm behavior choice request")

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

	resolutionID, ok := payloadMap["resolutionId"].(string)
	if !ok || resolutionID == "" {
		h.sendError(connection, "Missing resolutionId")
		return
	}

	choiceIndexFloat, ok := payloadMap["choiceIndex"].(float64)
	if !ok {
		log.Error("Missing or invalid choiceIndex")
		h.sendError(connection, "Missing or invalid choiceIndex")
		return
	}
	choiceIndex := int(choiceIndexFloat)

	var cardStorageTargets []string
	if targetsRaw, ok := payloadMap["cardStorageTargets"].([]interface{}); ok {
		for _, t := range targetsRaw {
			if s, ok := t.(string); ok {
				cardStorageTargets = append(cardStorageTargets, s)
			}
		}
	}

	log.Debug("Parsed confirm behavior choice request",
		slog.Int("choice_index", choiceIndex))

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, resolutionID, choiceIndex, cardStorageTargets)
	if err != nil {
		log.Error("Failed to execute confirm behavior choice action", slog.Any("error", err))
		connection.Send <- dto.WebSocketMessage{Type: dto.MessageTypeError, Payload: map[string]interface{}{"error": err.Error(), "resolutionId": resolutionID}}
		return
	}

	log.Debug("Behavior choice confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-behavior-choice",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmBehaviorChoiceHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
