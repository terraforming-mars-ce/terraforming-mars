package admin

import (
	"context"
	"fmt"
	"log/slog"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SetCurrentTurnAction handles the admin action to set the current turn
type SetCurrentTurnAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetCurrentTurnAction creates a new set current turn admin action
func NewSetCurrentTurnAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetCurrentTurnAction {
	return &SetCurrentTurnAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set current turn admin action
func (a *SetCurrentTurnAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("action", "admin_set_current_turn"),
	)
	log.Debug("Admin: Setting current turn")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if game.CurrentPhase() != shared.GamePhaseAction {
		return fmt.Errorf("turns can only be moved during the action phase")
	}

	target, err := game.GetPlayer(playerID)
	if err != nil {
		log.Warn("Player not found in game", slog.Any("error", err))
		return fmt.Errorf("player not found: %s", playerID)
	}
	if target.HasPassed() || target.HasExited() {
		return fmt.Errorf("player %s has passed or left", target.Name())
	}

	// Same allotment as a normal turn advance: unlimited for the last player still in.
	actions := 2
	if activeCount(game) == 1 {
		actions = -1
	}

	err = game.SetCurrentTurn(ctx, playerID, actions)
	if err != nil {
		log.Error("Failed to update current turn", slog.Any("error", err))
		return fmt.Errorf("failed to update current turn: %w", err)
	}

	log.Info("Admin set current turn completed", slog.Int("actions", actions))
	return nil
}

func activeCount(g *game.Game) int {
	count := 0
	for _, p := range g.GetAllPlayers() {
		if !p.HasPassed() && !p.HasExited() {
			count++
		}
	}
	return count
}
