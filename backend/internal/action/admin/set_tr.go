package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// SetTRAction handles the admin action to set player terraform rating
type SetTRAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetTRAction creates a new set TR admin action
func NewSetTRAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetTRAction {
	return &SetTRAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set TR admin action
func (a *SetTRAction) Execute(ctx context.Context, gameID string, playerID string, terraformRating int) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_tr"),
		zap.Int("terraform_rating", terraformRating),
	)
	log.Debug("Admin: Setting player terraform rating")

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

	player.Resources().SetTerraformRating(terraformRating)

	log.Info("Admin set terraform rating completed")
	return nil
}
