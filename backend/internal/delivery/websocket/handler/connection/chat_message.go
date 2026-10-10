package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game"
	"openmars/internal/game/shared"
	"openmars/internal/logger"
)

// ChatMessageHandler handles chat message requests from players and spectators.
type ChatMessageHandler struct {
	action      *connaction.SendChatMessageAction
	broadcaster ChatBroadcaster
	gameRepo    game.GameRepository
	logger      *slog.Logger
}

// ChatBroadcaster defines the broadcasting interface needed by the chat handler.
type ChatBroadcaster interface {
	BroadcastChatMessage(gameID string, chatMsg shared.ChatMessage)
}

// NewChatMessageHandler creates a new chat message handler.
func NewChatMessageHandler(
	action *connaction.SendChatMessageAction,
	broadcaster ChatBroadcaster,
	gameRepo game.GameRepository,
) *ChatMessageHandler {
	return &ChatMessageHandler{
		action:      action,
		broadcaster: broadcaster,
		gameRepo:    gameRepo,
		logger:      logger.Get(),
	}
}

// HandleMessage processes a chat-message from a player or spectator.
func (h *ChatMessageHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	gameID := connection.GameID()
	if gameID == "" {
		connection.SendError(message.Type, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "invalid payload format")
		return
	}

	msgText, _ := payloadMap["message"].(string)
	if msgText == "" {
		connection.SendError(message.Type, "message cannot be empty")
		return
	}

	isSpectator := connection.IsSpectator()
	var senderID, senderName, senderColor string

	g, err := h.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		connection.SendError(message.Type, "game not found")
		return
	}

	if isSpectator {
		s, err := g.GetSpectator(connection.SpectatorID())
		if err != nil {
			log.Error("Spectator not found", slog.Any("error", err))
			connection.SendError(message.Type, "spectator not found")
			return
		}
		senderID = s.ID()
		senderName = s.Name()
		senderColor = s.Color()
	} else {
		p, err := g.GetPlayer(connection.PlayerID())
		if err != nil {
			log.Error("Player not found", slog.Any("error", err))
			connection.SendError(message.Type, "player not found")
			return
		}
		senderID = p.ID()
		senderName = p.Name()
		senderColor = p.Color()
	}

	chatMsg, err := h.action.Execute(ctx, gameID, senderID, senderName, senderColor, msgText, isSpectator)
	if err != nil {
		log.Error("Failed to send chat message", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	h.broadcaster.BroadcastChatMessage(gameID, *chatMsg)
	log.Debug("Chat message broadcast")
}
