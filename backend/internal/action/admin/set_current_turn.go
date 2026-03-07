package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// SetCurrentTurnAction handles the admin action to set the current turn
type SetCurrentTurnAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetCurrentTurnAction creates a new set current turn admin action
func NewSetCurrentTurnAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetCurrentTurnAction {
	return &SetCurrentTurnAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set current turn admin action
func (a *SetCurrentTurnAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_current_turn"),
	)
	log.Debug("Admin: Setting current turn")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	_, err = game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	err = game.SetCurrentTurn(ctx, playerID, -1)
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update current turn: %w", err)
	}

	log.Info("Admin set current turn completed")
	return nil
}
