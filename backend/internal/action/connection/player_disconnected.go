package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// PlayerDisconnectedAction handles the business logic for player disconnection
type PlayerDisconnectedAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewPlayerDisconnectedAction creates a new player disconnected action
func NewPlayerDisconnectedAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlayerDisconnectedAction {
	return &PlayerDisconnectedAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the player disconnected action
func (a *PlayerDisconnectedAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "player_disconnected"),
	)
	log.Debug("Player disconnecting")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() == shared.GameStatusLobby {
		wasHost := g.HostPlayerID() == playerID

		if err := g.RemovePlayer(ctx, playerID); err != nil {
			log.Error("Failed to remove player from lobby", zap.Error(err))
			return fmt.Errorf("failed to remove player: %w", err)
		}

		remaining := g.GetAllPlayers()
		if len(remaining) == 0 {
			if err := a.gameRepo.Delete(ctx, gameID); err != nil {
				log.Error("Failed to delete empty game", zap.Error(err))
				return fmt.Errorf("failed to delete empty game: %w", err)
			}
			log.Debug("Game deleted (no players remaining)")
			return nil
		}

		if wasHost {
			var newHost string
			for _, p := range remaining {
				if !p.IsBot() {
					newHost = p.ID()
					break
				}
			}

			if newHost == "" {
				if err := a.gameRepo.Delete(ctx, gameID); err != nil {
					log.Error("Failed to delete game with only bots", zap.Error(err))
					return fmt.Errorf("failed to delete game: %w", err)
				}
				log.Debug("Game deleted (no human players remaining)")
				return nil
			}

			if err := g.SetHostPlayerID(ctx, newHost); err != nil {
				log.Error("Failed to reassign host", zap.Error(err))
				return fmt.Errorf("failed to reassign host: %w", err)
			}
			log.Debug("Host reassigned to human player", zap.String("new_host", newHost))
		}

		log.Debug("Player removed from lobby")
		return nil
	}

	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	if player.IsBot() {
		log.Debug("Skipping disconnect for bot player", zap.String("player_id", playerID))
		return nil
	}

	player.SetConnected(false)

	log.Info("Player disconnected")
	return nil
}
