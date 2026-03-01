package turn_management

import (
	"context"
	"fmt"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectStartingCardsAction handles the business logic for selecting starting project cards
type SelectStartingCardsAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the select starting cards action
func (a *SelectStartingCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_starting_cards"),
		zap.Strings("card_ids", cardIDs),
	)
	log.Info("🃏 Player selecting starting cards")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	selectionPhase := g.GetSelectStartingCardsPhase(playerID)
	if selectionPhase == nil {
		log.Error("Player not in starting card selection phase")
		return fmt.Errorf("not in starting card selection phase")
	}

	availableSet := make(map[string]bool)
	for _, id := range selectionPhase.AvailableCards {
		availableSet[id] = true
	}
	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	cost := len(cardIDs) * 3

	resources := player.Resources().Get()
	if resources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -cost,
	})

	log.Info("✅ Card selection cost deducted",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	baseaction.AddCardsToPlayerHand(cardIDs, player, g, a.cardRegistry, log)

	selectedSet := make(map[string]bool, len(cardIDs))
	for _, id := range cardIDs {
		selectedSet[id] = true
	}
	var unselectedProjectCards []string
	for _, cardID := range selectionPhase.AvailableCards {
		if !selectedSet[cardID] {
			unselectedProjectCards = append(unselectedProjectCards, cardID)
		}
	}
	if len(unselectedProjectCards) > 0 {
		if err := g.Deck().Discard(ctx, unselectedProjectCards); err != nil {
			log.Error("Failed to discard unselected project cards", zap.Error(err))
			return fmt.Errorf("failed to discard unselected project cards: %w", err)
		}
		log.Info("🗑️ Unselected project cards added to discard pile",
			zap.Int("count", len(unselectedProjectCards)))
	}

	if err := g.SetSelectStartingCardsPhase(ctx, playerID, nil); err != nil {
		log.Error("Failed to clear starting cards phase", zap.Error(err))
		return fmt.Errorf("failed to clear starting cards phase: %w", err)
	}
	log.Info("✅ Starting selection marked complete")

	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, p := range allPlayers {
		if g.GetSelectStartingCardsPhase(p.ID()) != nil {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed starting selection, advancing to action phase")

		if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			return fmt.Errorf("failed to transition game phase: %w", err)
		}

		turnOrder := g.TurnOrder()
		if len(turnOrder) > 0 {
			firstPlayerID := turnOrder[0]

			availableActions := 2
			if len(allPlayers) == 1 {
				availableActions = -1
				log.Info("🎮 Solo mode detected - setting unlimited actions")
			}

			if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}

			log.Info("✅ Set first player turn with actions",
				zap.String("first_player_id", firstPlayerID),
				zap.Int("available_actions", availableActions))
		}
	}

	log.Info("🎉 Starting card selection completed successfully")
	return nil
}
