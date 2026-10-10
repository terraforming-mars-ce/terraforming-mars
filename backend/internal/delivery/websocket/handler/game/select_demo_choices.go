package game

import (
	"context"
	"encoding/json"
	"log/slog"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game/shared"
	"openmars/internal/logger"
)

// SelectDemoChoicesHandler handles demo lobby card selection requests
type SelectDemoChoicesHandler struct {
	action      *gameaction.SelectDemoChoicesAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewSelectDemoChoicesHandler creates a new select demo choices handler
func NewSelectDemoChoicesHandler(action *gameaction.SelectDemoChoicesAction, broadcaster Broadcaster) *SelectDemoChoicesHandler {
	return &SelectDemoChoicesHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectDemoChoicesHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing select demo choices request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	var request dto.SelectDemoChoicesRequest
	if err := json.Unmarshal(payloadBytes, &request); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	choices := gameaction.DemoChoices{
		CorporationID: request.CorporationID,
		PreludeIDs:    request.PreludeIDs,
		CardIDs:       request.CardIDs,
		Resources: shared.Resources{
			Credits:  request.Resources.Credits,
			Steel:    request.Resources.Steel,
			Titanium: request.Resources.Titanium,
			Plants:   request.Resources.Plants,
			Energy:   request.Resources.Energy,
			Heat:     request.Resources.Heat,
		},
		Production: shared.Production{
			Credits:  request.Production.Credits,
			Steel:    request.Production.Steel,
			Titanium: request.Production.Titanium,
			Plants:   request.Production.Plants,
			Energy:   request.Production.Energy,
			Heat:     request.Production.Heat,
		},
		TerraformRating: request.TerraformRating,
		Generation:      request.Generation,
	}
	if request.GlobalParameters != nil {
		choices.GlobalParameters = &gameaction.DemoGlobalParameters{
			Temperature: request.GlobalParameters.Temperature,
			Oxygen:      request.GlobalParameters.Oxygen,
			Oceans:      request.GlobalParameters.Oceans,
		}
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, &choices)
	if err != nil {
		log.Error("Failed to execute select demo choices action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-demo-choices",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SelectDemoChoicesHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
