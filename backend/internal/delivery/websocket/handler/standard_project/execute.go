package standard_project

import (
	"context"
	"encoding/json"
	"log/slog"
	"openmars/internal/game/shared"
	"strings"

	stdprojaction "openmars/internal/action/standard_project"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// ExecuteHandler handles all standard project requests via a single unified handler
type ExecuteHandler struct {
	action      *stdprojaction.ExecuteStandardProjectAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewExecuteHandler creates a new unified standard project handler
func NewExecuteHandler(action *stdprojaction.ExecuteStandardProjectAction, broadcaster Broadcaster) *ExecuteHandler {
	return &ExecuteHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ExecuteHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	projectID := extractProjectID(message)
	if projectID == "" {
		log.Error("Missing projectId")
		connection.SendError(message.Type, "Missing projectId")
		return
	}

	log = log.With(slog.String("project_id", projectID))
	log.Debug("Processing standard project request")

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

	err := h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), projectID, paymentEnvelope.Payment)
	if err != nil {
		log.Error("Failed to execute standard project", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Standard project executed")

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}

// extractProjectID gets the project ID from either the payload or the legacy message type
func extractProjectID(message dto.WebSocketMessage) string {
	if payload, ok := message.Payload.(map[string]interface{}); ok {
		if projectID, ok := payload["projectId"].(string); ok && projectID != "" {
			return projectID
		}
	}

	// Legacy message type support: extract project ID from message type
	// e.g., "action.standard-project.sell-patents" -> "sell-patents"
	msgType := string(message.Type)
	return extractProjectIDFromMessageType(msgType)
}

// extractProjectIDFromMessageType maps legacy message types to project IDs
func extractProjectIDFromMessageType(msgType string) string {
	prefix := "action.standard-project."
	if !strings.HasPrefix(msgType, prefix) {
		return ""
	}
	suffix := msgType[len(prefix):]

	legacyMapping := map[string]string{
		"sell-patents":      "sell-patents",
		"build-power-plant": "power-plant",
		"launch-asteroid":   "asteroid",
		"build-aquifer":     "aquifer",
		"plant-greenery":    "greenery",
		"build-city":        "city",
	}

	if projectID, ok := legacyMapping[suffix]; ok {
		return projectID
	}
	return ""
}
