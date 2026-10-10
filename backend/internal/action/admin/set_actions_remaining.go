package admin

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
)

// SetActionsRemainingAction handles the admin action to set the current player's remaining actions
type SetActionsRemainingAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetActionsRemainingAction creates a new set actions remaining admin action
func NewSetActionsRemainingAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetActionsRemainingAction {
	return &SetActionsRemainingAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute sets the remaining actions of the player whose turn it is
func (a *SetActionsRemainingAction) Execute(ctx context.Context, gameID string, actions int) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.Int("actions", actions),
		slog.String("action", "admin_set_actions_remaining"),
	)
	log.Debug("Admin: Setting actions remaining")

	if actions < 1 {
		return fmt.Errorf("actions must be at least 1")
	}

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}
	if g.CurrentTurn() == nil || g.CurrentTurn().PlayerID() == "" {
		return fmt.Errorf("no current turn set")
	}

	if err := g.SetCurrentTurnActions(ctx, actions); err != nil {
		return fmt.Errorf("failed to set actions: %w", err)
	}

	log.Info("Admin set actions remaining")
	return nil
}
