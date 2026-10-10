package query

import (
	"context"
	"log/slog"

	"openmars/internal/game"
)

// GetGameAction handles querying a single game
type GetGameAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewGetGameAction creates a new get game query action
func NewGetGameAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *GetGameAction {
	return &GetGameAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute retrieves a game by ID
func (a *GetGameAction) Execute(ctx context.Context, gameID string) (*game.Game, error) {
	log := a.logger.With(slog.String("game_id", gameID))
	log.Debug("Querying game")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Debug("Failed to get game", slog.Any("error", err))
		return nil, err
	}

	log.Debug("Game query completed")
	return game, nil
}
