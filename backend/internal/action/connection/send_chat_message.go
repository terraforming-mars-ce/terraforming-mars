package connection

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"openmars/internal/game"
	"openmars/internal/game/shared"
)

// SendChatMessageAction handles sending a chat message in a game.
type SendChatMessageAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSendChatMessageAction creates a new SendChatMessageAction.
func NewSendChatMessageAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SendChatMessageAction {
	return &SendChatMessageAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute adds a chat message to the game.
func (a *SendChatMessageAction) Execute(ctx context.Context, gameID, senderID, senderName, senderColor, message string, isSpectator bool) (*shared.ChatMessage, error) {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("sender_name", senderName),
		slog.Bool("is_spectator", isSpectator),
		slog.String("action", "send_chat_message"),
	)

	if len(message) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	if len(message) > shared.MaxChatMessageLength {
		return nil, fmt.Errorf("message exceeds maximum length of %d characters", shared.MaxChatMessageLength)
	}

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	chatMsg := shared.ChatMessage{
		SenderID:    senderID,
		SenderName:  senderName,
		SenderColor: senderColor,
		Message:     message,
		Timestamp:   time.Now(),
		IsSpectator: isSpectator,
	}

	g.AddChatMessage(ctx, chatMsg)

	log.Debug("Chat message added")
	return &chatMsg, nil
}
