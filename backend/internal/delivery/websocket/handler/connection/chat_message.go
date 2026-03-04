package connection

import (
	"context"
	"time"

	connaction "terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ChatMessageHandler handles chat message requests from players and spectators.
type ChatMessageHandler struct {
	action      *connaction.SendChatMessageAction
	broadcaster ChatBroadcaster
	gameRepo    game.GameRepository
	logger      *zap.Logger
}

// ChatBroadcaster defines the broadcasting interface needed by the chat handler.
type ChatBroadcaster interface {
	BroadcastChatMessage(gameID string, chatMsg dto.ChatMessageDto)
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
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	gameID := connection.GameID
	if gameID == "" {
		h.sendError(connection, "not connected to game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "invalid payload format")
		return
	}

	msgText, _ := payloadMap["message"].(string)
	if msgText == "" {
		h.sendError(connection, "message cannot be empty")
		return
	}

	isSpectator := connection.IsSpectator()
	var senderID, senderName, senderColor string

	g, err := h.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		h.sendError(connection, "game not found")
		return
	}

	if isSpectator {
		s, err := g.GetSpectator(connection.SpectatorID)
		if err != nil {
			log.Error("Spectator not found", zap.Error(err))
			h.sendError(connection, "spectator not found")
			return
		}
		senderID = s.ID()
		senderName = s.Name()
		senderColor = s.Color()
	} else {
		p, err := g.GetPlayer(connection.PlayerID)
		if err != nil {
			log.Error("Player not found", zap.Error(err))
			h.sendError(connection, "player not found")
			return
		}
		senderID = p.ID()
		senderName = p.Name()
		senderColor = p.Color()
	}

	chatMsg, err := h.action.Execute(ctx, gameID, senderID, senderName, senderColor, msgText, isSpectator)
	if err != nil {
		log.Error("Failed to send chat message", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	chatDto := dto.ChatMessageDto{
		SenderID:    chatMsg.SenderID,
		SenderName:  chatMsg.SenderName,
		SenderColor: chatMsg.SenderColor,
		Message:     chatMsg.Message,
		Timestamp:   chatMsg.Timestamp.Format(time.RFC3339),
		IsSpectator: chatMsg.IsSpectator,
	}

	h.broadcaster.BroadcastChatMessage(gameID, chatDto)
	log.Debug("💬 Chat message broadcast")
}

func (h *ChatMessageHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.SendMessage(dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	})
}
