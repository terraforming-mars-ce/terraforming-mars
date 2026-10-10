package admin

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/board"
	"openmars/internal/game/shared"
)

// StartTileSelectionAction handles the admin action to start tile selection for a player
type StartTileSelectionAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewStartTileSelectionAction creates a new start tile selection admin action
func NewStartTileSelectionAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *StartTileSelectionAction {
	return &StartTileSelectionAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the start tile selection admin action
func (a *StartTileSelectionAction) Execute(ctx context.Context, gameID string, playerID string, tileType string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("action", "admin_start_tile_selection"),
		slog.String("tile_type", tileType),
	)
	log.Debug("Admin: Starting tile selection")

	if !board.ValidPlaceableTileType(tileType) {
		log.Error("Invalid tile type", slog.String("tile_type", tileType))
		return fmt.Errorf("invalid tile type: %s", tileType)
	}

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	_, err = g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", slog.Any("error", err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	queue := &shared.PendingTileSelectionQueue{
		Items:  []string{tileType},
		Source: "admin-tile-selection",
	}

	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		log.Error("Failed to set tile selection queue", slog.Any("error", err))
		return fmt.Errorf("failed to start tile selection: %w", err)
	}

	log.Info("Admin start tile selection completed")
	return nil
}
