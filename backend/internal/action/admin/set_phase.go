package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SetPhaseAction handles the admin action to set the game phase
type SetPhaseAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetPhaseAction creates a new set phase admin action
func NewSetPhaseAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetPhaseAction {
	return &SetPhaseAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set phase admin action
func (a *SetPhaseAction) Execute(ctx context.Context, gameID string, phase shared.GamePhase) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("action", "admin_set_phase"),
		zap.String("phase", string(phase)),
	)
	log.Debug("Admin: Setting game phase")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	err = game.UpdatePhase(ctx, phase)
	if err != nil {
		log.Error("Failed to update phase", zap.Error(err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("Admin set phase completed")
	return nil
}
