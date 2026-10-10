package game

import (
	"context"
	"log/slog"
	"time"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

type ConvertToBotHandler struct {
	action      *gameaction.ConvertToBotAction
	broadcaster Broadcaster
	hub         *core.Hub
	logger      *slog.Logger
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
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
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
		log.Error("Failed to convert player to bot", slog.Any("error", err))
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
