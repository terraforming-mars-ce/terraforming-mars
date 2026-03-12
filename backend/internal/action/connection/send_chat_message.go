package connection

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SendChatMessageAction handles sending a chat message in a game.
type SendChatMessageAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSendChatMessageAction creates a new SendChatMessageAction.
func NewSendChatMessageAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SendChatMessageAction {
	return &SendChatMessageAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute adds a chat message to the game.
func (a *SendChatMessageAction) Execute(ctx context.Context, gameID, senderID, senderName, senderColor, message string, isSpectator bool) (*shared.ChatMessage, error) {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("sender_name", senderName),
		zap.Bool("is_spectator", isSpectator),
		zap.String("action", "send_chat_message"),
	)

	if len(message) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	if len(message) > shared.MaxChatMessageLength {
		return nil, fmt.Errorf("message exceeds maximum length of %d characters", shared.MaxChatMessageLength)
	}

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
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
