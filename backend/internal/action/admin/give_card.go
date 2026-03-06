package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
)

// GiveCardAction handles the admin action to give a card to a player
// NOTE: Card validation is skipped (admin action with trusted input)
type GiveCardAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewGiveCardAction creates a new give card admin action
func NewGiveCardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *GiveCardAction {
	return &GiveCardAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the give card admin action
func (a *GiveCardAction) Execute(ctx context.Context, gameID string, playerID string, cardID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_give_card"),
		zap.String("card_id", cardID),
	)
	log.Debug("Admin: Giving card to player")

	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	action.AddCardsToPlayerHand([]string{cardID}, player, game, a.cardRegistry, log)

	log.Info("Admin give card completed")
	return nil
}
