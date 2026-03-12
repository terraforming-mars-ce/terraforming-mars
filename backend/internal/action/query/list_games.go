package query

import (
	"context"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ListGamesAction handles querying all games
type ListGamesAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewListGamesAction creates a new list games query action
func NewListGamesAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *ListGamesAction {
	return &ListGamesAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute retrieves all games, optionally filtered by status
func (a *ListGamesAction) Execute(ctx context.Context, status *shared.GameStatus) ([]*game.Game, error) {
	log := a.logger
	log.Debug("Querying all games")

	games, err := a.gameRepo.List(ctx, status)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err))
		return nil, err
	}

	log.Debug("Games query completed", zap.Int("count", len(games)))
	return games, nil
}
