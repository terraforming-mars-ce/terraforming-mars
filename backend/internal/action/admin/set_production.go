package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SetProductionAction handles the admin action to set player production
type SetProductionAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetProductionAction creates a new set production admin action
func NewSetProductionAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetProductionAction {
	return &SetProductionAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set production admin action
func (a *SetProductionAction) Execute(ctx context.Context, gameID string, playerID string, production shared.Production) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_production"),
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat),
	)
	log.Debug("Admin: Setting player production")

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

	player.Resources().SetProduction(production)

	log.Info("Admin set production completed")
	return nil
}
