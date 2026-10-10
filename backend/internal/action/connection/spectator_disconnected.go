package connection

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
)

// SpectatorDisconnectedAction handles spectator disconnection by removing them from the game.
type SpectatorDisconnectedAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSpectatorDisconnectedAction creates a new SpectatorDisconnectedAction.
func NewSpectatorDisconnectedAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SpectatorDisconnectedAction {
	return &SpectatorDisconnectedAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute removes a spectator from the game entirely (spectators are ephemeral).
func (a *SpectatorDisconnectedAction) Execute(ctx context.Context, gameID, spectatorID string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("spectator_id", spectatorID),
		slog.String("action", "spectator_disconnected"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Debug("Game not found for spectator disconnect (may be deleted)", slog.Any("error", err))
		return nil
	}

	if err := g.RemoveSpectator(ctx, spectatorID); err != nil {
		log.Debug("Spectator not found during disconnect", slog.Any("error", err))
		return nil
	}

	log.Debug("Spectator disconnected and removed")
	return nil
}

// KickSpectatorAction handles kicking a spectator from a game.
type KickSpectatorAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewKickSpectatorAction creates a new KickSpectatorAction.
func NewKickSpectatorAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *KickSpectatorAction {
	return &KickSpectatorAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute kicks a spectator from the game. Only the host can kick spectators.
func (a *KickSpectatorAction) Execute(ctx context.Context, gameID, requesterID, targetSpectatorID string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("requester_id", requesterID),
		slog.String("target_spectator_id", targetSpectatorID),
		slog.String("action", "kick_spectator"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("only the host can kick spectators")
	}

	if err := g.RemoveSpectator(ctx, targetSpectatorID); err != nil {
		log.Error("Failed to remove spectator", slog.Any("error", err))
		return fmt.Errorf("spectator not found: %s", targetSpectatorID)
	}

	log.Info("Spectator kicked from game")
	return nil
}
