package query

import (
	"context"

	"terraforming-mars-backend/internal/game"

	"go.uber.org/zap"
)

// GetGameAction handles querying a single game
type GetGameAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewGetGameAction creates a new get game query action
func NewGetGameAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *GetGameAction {
	return &GetGameAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute retrieves a game by ID
func (a *GetGameAction) Execute(ctx context.Context, gameID string) (*game.Game, error) {
	log := a.logger.With(zap.String("game_id", gameID))
	log.Debug("Querying game")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Debug("Failed to get game", zap.Error(err))
		return nil, err
	}

	log.Debug("Game query completed")
	return game, nil
}
