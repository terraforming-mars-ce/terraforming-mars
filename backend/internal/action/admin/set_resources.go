package admin

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/shared"
)

// SetResourcesAction handles the admin action to set player resources
type SetResourcesAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetResourcesAction creates a new set resources admin action
func NewSetResourcesAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetResourcesAction {
	return &SetResourcesAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set resources admin action
func (a *SetResourcesAction) Execute(ctx context.Context, gameID string, playerID string, resources shared.Resources) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("action", "admin_set_resources"),
		slog.Int("credits", resources.Credits),
		slog.Int("steel", resources.Steel),
		slog.Int("titanium", resources.Titanium),
		slog.Int("plants", resources.Plants),
		slog.Int("energy", resources.Energy),
		slog.Int("heat", resources.Heat),
	)
	log.Debug("Admin: Setting player resources")

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

	player.Resources().Set(resources)

	log.Info("Admin set resources completed")
	return nil
}
