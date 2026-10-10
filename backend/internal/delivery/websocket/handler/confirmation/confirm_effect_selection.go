package confirmation

import (
	"context"
	"log/slog"

	confirmaction "openmars/internal/action/confirmation"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// ConfirmEffectSelectionHandler handles confirm effect selection requests
type ConfirmEffectSelectionHandler struct {
	action      *confirmaction.ConfirmEffectSelectionAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewConfirmEffectSelectionHandler creates a new confirm effect selection handler
func NewConfirmEffectSelectionHandler(action *confirmaction.ConfirmEffectSelectionAction, broadcaster Broadcaster) *ConfirmEffectSelectionHandler {
	return &ConfirmEffectSelectionHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmEffectSelectionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm effect selection request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		h.sendError(connection, "Invalid payload format")
		return
	}

	option, ok := payloadMap["optionIndex"].(float64)
	if !ok || option != float64(int(option)) {
		h.sendError(connection, "Missing or invalid optionIndex in payload")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, int(option))
	if err != nil {
		log.Error("Failed to execute confirm effect selection action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Effect selection confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-effect-selection",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmEffectSelectionHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
