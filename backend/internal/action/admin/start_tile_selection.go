package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
)

// StartTileSelectionAction handles the admin action to start tile selection for a player
type StartTileSelectionAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewStartTileSelectionAction creates a new start tile selection admin action
func NewStartTileSelectionAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *StartTileSelectionAction {
	return &StartTileSelectionAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the start tile selection admin action
func (a *StartTileSelectionAction) Execute(ctx context.Context, gameID string, playerID string, tileType string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_start_tile_selection"),
		zap.String("tile_type", tileType),
	)
	log.Debug("Admin: Starting tile selection")

	if !board.ValidPlaceableTileType(tileType) {
		log.Error("Invalid tile type", zap.String("tile_type", tileType))
		return fmt.Errorf("invalid tile type: %s", tileType)
	}

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	_, err = g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	queue := &shared.PendingTileSelectionQueue{
		Items:  []string{tileType},
		Source: "admin-tile-selection",
	}

	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		log.Error("Failed to set tile selection queue", zap.Error(err))
		return fmt.Errorf("failed to start tile selection: %w", err)
	}

	log.Info("Admin start tile selection completed")
	return nil
}
