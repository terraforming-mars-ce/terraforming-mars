package game

import (
	"context"
	"time"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

type ConvertToBotHandler struct {
	action      *gameaction.ConvertToBotAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *zap.Logger
}

func NewConvertToBotHandler(action *gameaction.ConvertToBotAction, broadcaster Broadcaster, hub *core.Hub) *ConvertToBotHandler {
	return &ConvertToBotHandler{
		action:      action,
		broadcaster: broadcaster,
		hub:         hub,
		logger:      logger.Get(),
	}
}

func (h *ConvertToBotHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)
	log.Debug("Processing convert to bot request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]any)
	if !ok {
		h.sendError(connection, "invalid payload format")
		return
	}

	targetPlayerID, _ := payloadMap["targetPlayerId"].(string)
	if targetPlayerID == "" {
		h.sendError(connection, "targetPlayerId is required")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, targetPlayerID)
	if err != nil {
		log.Error("Failed to convert player to bot", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Player converted to bot")

	// Send player-kicked to the converted player so they return to main menu
	convertedConn := h.hub.GetManager().GetConnectionByPlayerID(connection.GameID, targetPlayerID)
	if convertedConn != nil {
		kickMsg := dto.WebSocketMessage{
			Type:    dto.MessageTypePlayerKicked,
			GameID:  connection.GameID,
			Payload: map[string]any{"reason": "You were replaced by a bot"},
		}
		convertedConn.SendMessage(kickMsg)

		go func() {
			time.Sleep(100 * time.Millisecond)
			convertedConn.Close()
		}()
	}

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
}

func (h *ConvertToBotHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]any{
			"error": errorMessage,
		},
	}
}
