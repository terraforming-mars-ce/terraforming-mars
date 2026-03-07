package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// SetGlobalParametersRequest contains the parameters to set
type SetGlobalParametersRequest struct {
	Temperature int
	Oxygen      int
	Oceans      int
	Venus       int
}

// SetGlobalParametersAction handles the admin action to set global parameters
type SetGlobalParametersAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetGlobalParametersAction creates a new set global parameters admin action
func NewSetGlobalParametersAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetGlobalParametersAction {
	return &SetGlobalParametersAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set global parameters admin action
func (a *SetGlobalParametersAction) Execute(ctx context.Context, gameID string, params SetGlobalParametersRequest) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("action", "admin_set_global_parameters"),
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans),
		zap.Int("venus", params.Venus),
	)
	log.Debug("Admin: Setting global parameters")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if err := game.GlobalParameters().SetTemperature(ctx, params.Temperature); err != nil {
		log.Error("Failed to update temperature", zap.Error(err))
		return fmt.Errorf("failed to update temperature: %w", err)
	}

	if err := game.GlobalParameters().SetOxygen(ctx, params.Oxygen); err != nil {
		log.Error("Failed to update oxygen", zap.Error(err))
		return fmt.Errorf("failed to update oxygen: %w", err)
	}

	if err := game.GlobalParameters().SetOceans(ctx, params.Oceans); err != nil {
		log.Error("Failed to update oceans", zap.Error(err))
		return fmt.Errorf("failed to update oceans: %w", err)
	}

	if err := game.GlobalParameters().SetVenus(ctx, params.Venus); err != nil {
		log.Error("Failed to update venus", zap.Error(err))
		return fmt.Errorf("failed to update venus: %w", err)
	}

	log.Info("Admin set global parameters completed")
	return nil
}
