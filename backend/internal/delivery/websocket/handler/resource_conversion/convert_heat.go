package resource_conversion

import (
	"context"
	"encoding/json"

	resconvaction "terraforming-mars-backend/internal/action/resource_conversion"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConvertHeatHandler handles convert heat to temperature requests
type ConvertHeatHandler struct {
	action      *resconvaction.ConvertHeatToTemperatureAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewConvertHeatHandler creates a new convert heat handler
func NewConvertHeatHandler(action *resconvaction.ConvertHeatToTemperatureAction, broadcaster Broadcaster) *ConvertHeatHandler {
	return &ConvertHeatHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConvertHeatHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing convert heat request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	var req dto.ActionConvertHeatToTemperatureRequest
	if message.Payload != nil {
		payloadBytes, _ := json.Marshal(message.Payload)
		_ = json.Unmarshal(payloadBytes, &req)
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, req.StorageSubstitutes)
	if err != nil {
		log.Error("Failed to execute convert heat action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Heat conversion completed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "convert-heat",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConvertHeatHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
