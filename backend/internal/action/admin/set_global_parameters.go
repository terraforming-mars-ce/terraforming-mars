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
	log.Info("🌍 Admin: Setting global parameters")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if params.Temperature != 0 {
		err := game.GlobalParameters().SetTemperature(ctx, params.Temperature)
		if err != nil {
			log.Error("Failed to update temperature", zap.Error(err))
			return fmt.Errorf("failed to update temperature: %w", err)
		}
	}

	if params.Oxygen != 0 {
		err := game.GlobalParameters().SetOxygen(ctx, params.Oxygen)
		if err != nil {
			log.Error("Failed to update oxygen", zap.Error(err))
			return fmt.Errorf("failed to update oxygen: %w", err)
		}
	}

	if params.Oceans != 0 {
		err := game.GlobalParameters().SetOceans(ctx, params.Oceans)
		if err != nil {
			log.Error("Failed to update oceans", zap.Error(err))
			return fmt.Errorf("failed to update oceans: %w", err)
		}
	}

	if params.Venus != 0 {
		err := game.GlobalParameters().SetVenus(ctx, params.Venus)
		if err != nil {
			log.Error("Failed to update venus", zap.Error(err))
			return fmt.Errorf("failed to update venus: %w", err)
		}
	}

	log.Info("✅ Admin set global parameters completed")
	return nil
}
