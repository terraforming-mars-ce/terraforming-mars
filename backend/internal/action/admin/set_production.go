package admin

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/shared"
)

// SetProductionAction handles the admin action to set player production
type SetProductionAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetProductionAction creates a new set production admin action
func NewSetProductionAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetProductionAction {
	return &SetProductionAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set production admin action
func (a *SetProductionAction) Execute(ctx context.Context, gameID string, playerID string, production shared.Production) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("action", "admin_set_production"),
		slog.Int("credits", production.Credits),
		slog.Int("steel", production.Steel),
		slog.Int("titanium", production.Titanium),
		slog.Int("plants", production.Plants),
		slog.Int("energy", production.Energy),
		slog.Int("heat", production.Heat),
	)
	log.Debug("Admin: Setting player production")

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

	player.Resources().SetProduction(production)

	log.Info("Admin set production completed")
	return nil
}
