package query

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"

	"go.uber.org/zap"
)

// GetPlayerAction handles querying a single player
type GetPlayerAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *GetPlayerAction {
	return &GetPlayerAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute retrieves a player from a game
func (a *GetPlayerAction) Execute(ctx context.Context, gameID string, playerID string) (*player.Player, error) {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
	log.Debug("Querying player")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return nil, err
	}

	// Get player from game
	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	log.Debug("Player query completed")
	return player, nil
}
