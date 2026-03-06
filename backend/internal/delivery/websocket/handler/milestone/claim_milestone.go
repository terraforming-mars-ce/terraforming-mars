package milestone

import (
	"context"
	"encoding/json"

	milestoneaction "terraforming-mars-backend/internal/action/milestone"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster defines the interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// ClaimMilestoneHandler handles claim milestone requests
type ClaimMilestoneHandler struct {
	action      *milestoneaction.ClaimMilestoneAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewClaimMilestoneHandler creates a new claim milestone handler
func NewClaimMilestoneHandler(action *milestoneaction.ClaimMilestoneAction, broadcaster Broadcaster) *ClaimMilestoneHandler {
	return &ClaimMilestoneHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// ClaimMilestonePayload represents the expected payload for claiming a milestone
type ClaimMilestonePayload struct {
	MilestoneType string `json:"milestoneType"`
}

// HandleMessage implements the MessageHandler interface
func (h *ClaimMilestoneHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing claim milestone request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", zap.Error(err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	var payload ClaimMilestonePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", zap.Error(err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	if payload.MilestoneType == "" {
		log.Error("Missing milestone type in payload")
		h.sendError(connection, "Milestone type is required")
		return
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, payload.MilestoneType)
	if err != nil {
		log.Error("Failed to execute claim milestone action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Milestone claimed",
		zap.String("milestone_type", payload.MilestoneType))

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":        "claim-milestone",
			"milestoneType": payload.MilestoneType,
			"success":       true,
		},
	}

	connection.Send <- response
}

// sendError sends an error message to the client
func (h *ClaimMilestoneHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
