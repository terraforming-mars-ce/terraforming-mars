package projectfunding

import (
	"context"
	"encoding/json"
	"log/slog"

	pfAction "openmars/internal/action/projectfunding"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
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
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing project funding seat purchase")

	if connection.GameID() == "" || connection.PlayerID() == "" {
		log.Error("Missing connection context")
		connection.SendError(message.Type, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", slog.Any("error", err))
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	var payload FundSeatPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	if payload.ProjectID == "" {
		log.Error("Missing project ID in payload")
		connection.SendError(message.Type, "Project ID is required")
		return
	}

	payment := pfAction.FundSeatPayment{
		Credits:  payload.Credits,
		Steel:    payload.Steel,
		Titanium: payload.Titanium,
	}

	err = h.action.Execute(ctx, connection.GameID(), connection.PlayerID(), payload.ProjectID, payment)
	if err != nil {
		log.Error("Failed to execute fund seat action", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Project seat purchased", slog.String("project_id", payload.ProjectID))

	h.broadcaster.BroadcastGameState(connection.GameID(), nil)

}
