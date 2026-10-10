package connection

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/service/bot"
)

// EndGameAction handles ending a game and cleaning up all resources.
type EndGameAction struct {
	gameRepo       game.GameRepository
	botGameStopper bot.BotGameStopper
	logger         *slog.Logger
}

// NewEndGameAction creates a new EndGameAction.
func NewEndGameAction(
	gameRepo game.GameRepository,
	botGameStopper bot.BotGameStopper,
	logger *slog.Logger,
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
		slog.String("game_id", gameID),
		slog.String("requester_id", requesterID),
		slog.String("action", "end_game"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("cannot end game: only the host can end the game")
	}

	if a.botGameStopper != nil {
		go a.botGameStopper.StopAllBotsForGame(gameID)
	}

	if err := a.gameRepo.Delete(ctx, gameID); err != nil {
		log.Error("Failed to delete game", slog.Any("error", err))
		return fmt.Errorf("failed to delete game: %w", err)
	}

	log.Info("Game ended and deleted by host")
	return nil
}
