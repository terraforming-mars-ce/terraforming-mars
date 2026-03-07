package game

import (
	"context"
	"encoding/json"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConfirmDemoSetupHandler handles confirm demo setup requests
type ConfirmDemoSetupHandler struct {
	action      *gameaction.ConfirmDemoSetupAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewConfirmDemoSetupHandler creates a new confirm demo setup handler
func NewConfirmDemoSetupHandler(action *gameaction.ConfirmDemoSetupAction, broadcaster Broadcaster) *ConfirmDemoSetupHandler {
	return &ConfirmDemoSetupHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ConfirmDemoSetupHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing confirm demo setup request")

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

	var request dto.ConfirmDemoSetupRequest
	if err := json.Unmarshal(payloadBytes, &request); err != nil {
		log.Error("Failed to unmarshal payload", zap.Error(err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	log.Debug("Parsed confirm demo setup request",
		zap.Stringp("corporation_id", request.CorporationID),
		zap.Strings("card_ids", request.CardIDs),
		zap.Int("terraform_rating", request.TerraformRating))

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, &request)
	if err != nil {
		log.Error("Failed to execute confirm demo setup action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Demo setup confirmed")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "confirm-demo-setup",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *ConfirmDemoSetupHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
