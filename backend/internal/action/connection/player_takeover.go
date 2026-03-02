package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
)

// PlayerTakeoverAction handles the business logic for taking over a disconnected player
type PlayerTakeoverAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// PlayerTakeoverResult contains the result of a player takeover
type PlayerTakeoverResult struct {
	PlayerID   string
	PlayerName string
	GameDto    dto.GameDto
}

// NewPlayerTakeoverAction creates a new player takeover action
func NewPlayerTakeoverAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
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
		zap.String("game_id", gameID),
		zap.String("target_player_id", targetPlayerID),
		zap.String("action", "player_takeover"),
	)
	log.Info("🔄 Processing player takeover request")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	player, err := g.GetPlayer(targetPlayerID)
	if err != nil {
		log.Error("Target player not found in game", zap.Error(err))
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

	gameDto := dto.ToGameDto(g, a.cardRegistry, targetPlayerID)

	log.Info("✅ Player takeover completed successfully",
		zap.String("player_name", player.Name()))

	return &PlayerTakeoverResult{
		PlayerID:   targetPlayerID,
		PlayerName: player.Name(),
		GameDto:    gameDto,
	}, nil
}
