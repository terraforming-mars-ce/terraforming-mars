package projectfunding

import (
	"context"
	"encoding/json"

	pfAction "terraforming-mars-backend/internal/action/projectfunding"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster is the interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// FundSeatPayload represents the expected payload for buying a project seat
type FundSeatPayload struct {
	ProjectID string `json:"projectId"`
	Credits   int    `json:"credits"`
	Steel     int    `json:"steel"`
	Titanium  int    `json:"titanium"`
}

// FundSeatHandler handles project funding seat purchase requests
type FundSeatHandler struct {
	action      *pfAction.FundSeatAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewFundSeatHandler creates a new fund seat handler
func NewFundSeatHandler(action *pfAction.FundSeatAction, broadcaster Broadcaster) *FundSeatHandler {
	return &FundSeatHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *FundSeatHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing project funding seat purchase")

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

	var payload FundSeatPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", zap.Error(err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	if payload.ProjectID == "" {
		log.Error("Missing project ID in payload")
		h.sendError(connection, "Project ID is required")
		return
	}

	payment := pfAction.FundSeatPayment{
		Credits:  payload.Credits,
		Steel:    payload.Steel,
		Titanium: payload.Titanium,
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, payload.ProjectID, payment)
	if err != nil {
		log.Error("Failed to execute fund seat action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Project seat purchased", zap.String("project_id", payload.ProjectID))

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	connection.Send <- dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":    "buy-project-seat",
			"projectId": payload.ProjectID,
			"success":   true,
		},
	}
}

func (h *FundSeatHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
