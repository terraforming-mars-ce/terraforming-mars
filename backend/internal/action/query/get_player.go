package query

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/player"
)

// GetPlayerAction handles querying a single player
type GetPlayerAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *GetPlayerAction {
	return &GetPlayerAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute retrieves a player from a game
func (a *GetPlayerAction) Execute(ctx context.Context, gameID string, playerID string) (*player.Player, error) {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
	)
	log.Debug("Querying player")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return nil, err
	}

	// Get player from game
	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", slog.Any("error", err))
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	log.Debug("Player query completed")
	return player, nil
}
