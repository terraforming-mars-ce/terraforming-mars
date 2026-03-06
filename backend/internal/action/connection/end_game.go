package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/service/bot"
)

// EndGameAction handles ending a game and cleaning up all resources.
type EndGameAction struct {
	gameRepo       game.GameRepository
	botGameStopper bot.BotGameStopper
	logger         *zap.Logger
}

// NewEndGameAction creates a new EndGameAction.
func NewEndGameAction(
	gameRepo game.GameRepository,
	botGameStopper bot.BotGameStopper,
	logger *zap.Logger,
) *EndGameAction {
	return &EndGameAction{
		gameRepo:       gameRepo,
		botGameStopper: botGameStopper,
		logger:         logger,
	}
}

// Execute ends a game. Only the host can end the game.
func (a *EndGameAction) Execute(ctx context.Context, gameID string, requesterID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("requester_id", requesterID),
		zap.String("action", "end_game"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("cannot end game: only the host can end the game")
	}

	if a.botGameStopper != nil {
		go a.botGameStopper.StopAllBotsForGame(gameID)
	}

	if err := a.gameRepo.Delete(ctx, gameID); err != nil {
		log.Error("Failed to delete game", zap.Error(err))
		return fmt.Errorf("failed to delete game: %w", err)
	}

	log.Info("Game ended and deleted by host")
	return nil
}
