package admin

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
)

// SetTRAction handles the admin action to set player terraform rating
type SetTRAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetTRAction creates a new set TR admin action
func NewSetTRAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetTRAction {
	return &SetTRAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set TR admin action
func (a *SetTRAction) Execute(ctx context.Context, gameID string, playerID string, terraformRating int) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("action", "admin_set_tr"),
		slog.Int("terraform_rating", terraformRating),
	)
	log.Debug("Admin: Setting player terraform rating")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", slog.Any("error", err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	player.Resources().SetTerraformRating(terraformRating)

	log.Info("Admin set terraform rating completed")
	return nil
}
