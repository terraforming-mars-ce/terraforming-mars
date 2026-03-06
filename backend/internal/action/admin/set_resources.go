package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SetResourcesAction handles the admin action to set player resources
type SetResourcesAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetResourcesAction creates a new set resources admin action
func NewSetResourcesAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetResourcesAction {
	return &SetResourcesAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set resources admin action
func (a *SetResourcesAction) Execute(ctx context.Context, gameID string, playerID string, resources shared.Resources) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_resources"),
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat),
	)
	log.Debug("Admin: Setting player resources")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	player.Resources().Set(resources)

	log.Info("Admin set resources completed")
	return nil
}
