package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// SpectateGameAction handles a spectator joining a game.
type SpectateGameAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSpectateGameAction creates a new SpectateGameAction.
func NewSpectateGameAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SpectateGameAction {
	return &SpectateGameAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// SpectateGameResult contains the result of a spectator joining.
type SpectateGameResult struct {
	SpectatorID string
}

// Execute adds a spectator to the game.
func (a *SpectateGameAction) Execute(ctx context.Context, gameID, spectatorName, spectatorID string) (*SpectateGameResult, error) {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("spectator_name", spectatorName),
		zap.String("spectator_id", spectatorID),
		zap.String("action", "spectate_game"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if g.SpectatorCount() >= game.MaxSpectators {
		return nil, fmt.Errorf("game %s already has the maximum number of spectators (%d)", gameID, game.MaxSpectators)
	}

	color := g.NextSpectatorColor()
	spectator := game.NewSpectator(spectatorID, spectatorName, color)

	if err := g.AddSpectator(ctx, spectator); err != nil {
		log.Error("Failed to add spectator", zap.Error(err))
		return nil, fmt.Errorf("failed to add spectator: %w", err)
	}

	log.Info("Spectator joined game")

	return &SpectateGameResult{
		SpectatorID: spectatorID,
	}, nil
}
