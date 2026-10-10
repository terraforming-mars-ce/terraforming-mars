package connection

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/cards"
)

// PlayerTakeoverAction handles the business logic for taking over a disconnected player
type PlayerTakeoverAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *slog.Logger
}

// PlayerTakeoverResult contains the result of a player takeover
type PlayerTakeoverResult struct {
	PlayerID   string
	PlayerName string
}

// NewPlayerTakeoverAction creates a new player takeover action
func NewPlayerTakeoverAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *slog.Logger,
) *PlayerTakeoverAction {
	return &PlayerTakeoverAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the player takeover action
func (a *PlayerTakeoverAction) Execute(ctx context.Context, gameID string, targetPlayerID string) (*PlayerTakeoverResult, error) {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("target_player_id", targetPlayerID),
		slog.String("action", "player_takeover"),
	)
	log.Debug("Processing player takeover request")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if g.ResumeLobby() != nil {
		return nil, fmt.Errorf("use the resume lobby to join this game")
	}
	player, err := g.GetPlayer(targetPlayerID)
	if err != nil {
		log.Error("Target player not found in game", slog.Any("error", err))
		return nil, fmt.Errorf("player not found: %s", targetPlayerID)
	}

	if player.HasExited() {
		log.Warn("Cannot take over an exited player")
		return nil, fmt.Errorf("player has been kicked from the game")
	}

	if player.IsBot() {
		log.Warn("Cannot take over a bot player")
		return nil, fmt.Errorf("cannot take over a bot player")
	}

	if player.IsConnected() {
		log.Warn("Cannot take over a connected player")
		return nil, fmt.Errorf("player is already connected")
	}

	player.SetConnected(true)

	log.Info("Player takeover completed",
		slog.String("player_name", player.Name()))

	return &PlayerTakeoverResult{
		PlayerID:   targetPlayerID,
		PlayerName: player.Name(),
	}, nil
}
