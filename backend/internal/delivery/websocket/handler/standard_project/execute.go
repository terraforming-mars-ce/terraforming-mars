package standard_project

import (
	"context"
	"strings"

	stdprojaction "terraforming-mars-backend/internal/action/standard_project"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// ExecuteHandler handles all standard project requests via a single unified handler
type ExecuteHandler struct {
	action      *stdprojaction.ExecuteStandardProjectAction
	broadcaster Broadcaster
	logger      *zap.Logger
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	projectID := extractProjectID(message)
	if projectID == "" {
		log.Error("Missing projectId")
		h.sendError(connection, "Missing projectId")
		return
	}

	log = log.With(zap.String("project_id", projectID))
	log.Debug("Processing standard project request")

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, projectID)
	if err != nil {
		log.Error("Failed to execute standard project", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Standard project executed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":    "standard-project",
			"projectId": projectID,
			"success":   true,
		},
	}
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

func (h *ExecuteHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
