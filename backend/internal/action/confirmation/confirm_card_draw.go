package confirmation

import (
	"context"
	"fmt"
	"slices"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ConfirmCardDrawAction handles the business logic for confirming card draw selection
type ConfirmCardDrawAction struct {
	baseaction.BaseAction
}

// NewConfirmCardDrawAction creates a new confirm card draw action
func NewConfirmCardDrawAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmCardDrawAction {
	return &ConfirmCardDrawAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute performs the confirm card draw action
func (a *ConfirmCardDrawAction) Execute(ctx context.Context, gameID string, playerID string, cardsToTake []string, cardsToBuy []string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_card_draw"),
		zap.Int("cards_to_take", len(cardsToTake)),
		zap.Int("cards_to_buy", len(cardsToBuy)),
	)
	log.Info("🃏 Confirming card draw selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	selection := player.Selection().GetPendingCardDrawSelection()
	if selection == nil {
		log.Warn("No pending card draw selection found")
		return fmt.Errorf("no pending card draw selection found")
	}

	totalSelected := len(cardsToTake) + len(cardsToBuy)
	maxAllowed := selection.FreeTakeCount + selection.MaxBuyCount

	if totalSelected > maxAllowed {
		log.Warn("Too many cards selected",
			zap.Int("selected", totalSelected),
			zap.Int("max_allowed", maxAllowed))
		return fmt.Errorf("too many cards selected: selected %d, max allowed %d", totalSelected, maxAllowed)
	}

	if len(cardsToTake) > selection.FreeTakeCount {
		log.Warn("Too many free cards selected",
			zap.Int("selected", len(cardsToTake)),
			zap.Int("max", selection.FreeTakeCount))
		return fmt.Errorf("too many free cards selected: selected %d, max %d", len(cardsToTake), selection.FreeTakeCount)
	}

	isPureCardDraw := selection.MaxBuyCount == 0 && selection.FreeTakeCount == len(selection.AvailableCards)
	if isPureCardDraw && len(cardsToTake) != selection.FreeTakeCount {
		log.Warn("Must take all cards for pure card-draw effect",
			zap.Int("required", selection.FreeTakeCount),
			zap.Int("selected", len(cardsToTake)))
		return fmt.Errorf("must take all %d cards for card-draw effect", selection.FreeTakeCount)
	}

	if len(cardsToBuy) > selection.MaxBuyCount {
		log.Warn("Too many cards to buy",
			zap.Int("selected", len(cardsToBuy)),
			zap.Int("max", selection.MaxBuyCount))
		return fmt.Errorf("too many cards to buy: selected %d, max %d", len(cardsToBuy), selection.MaxBuyCount)
	}

	allSelectedCards := append(cardsToTake, cardsToBuy...)
	for _, cardID := range allSelectedCards {
		if !slices.Contains(selection.AvailableCards, cardID) {
			log.Warn("Card not in available cards", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	totalCost := len(cardsToBuy) * selection.CardBuyCost

	if totalCost > 0 {
		resources := player.Resources().Get()
		if resources.Credits < totalCost {
			log.Warn("Insufficient credits to buy cards",
				zap.Int("needed", totalCost),
				zap.Int("available", resources.Credits))
			return fmt.Errorf("insufficient credits to buy cards: need %d, have %d", totalCost, resources.Credits)
		}

		player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -totalCost,
		})

		newResources := player.Resources().Get()
		log.Info("💰 Paid for bought cards",
			zap.Int("cards_bought", len(cardsToBuy)),
			zap.Int("cost", totalCost),
			zap.Int("remaining_credits", newResources.Credits))
	}

	baseaction.AddCardsToPlayerHand(allSelectedCards, player, g, a.CardRegistry(), log)

	log.Info("🃏 Added selected cards to hand",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	unselectedCards := []string{}
	for _, cardID := range selection.AvailableCards {
		if !slices.Contains(allSelectedCards, cardID) {
			unselectedCards = append(unselectedCards, cardID)
		}
	}

	if len(unselectedCards) > 0 {
		if err := g.Deck().Discard(ctx, unselectedCards); err != nil {
			log.Error("Failed to discard unselected cards", zap.Error(err))
			return fmt.Errorf("failed to discard unselected cards: %w", err)
		}
		log.Debug("🗑️ Discarded unselected cards to discard pile",
			zap.Int("count", len(unselectedCards)),
			zap.Strings("card_ids", unselectedCards))
	}

	player.Selection().SetPendingCardDrawSelection(nil)

	// If this selection was triggered by a card action, complete the action now
	if selection.SourceCardID != "" {
		a.completeSourceCardAction(g, player, selection, log)
	}

	log.Info("✅ Card draw confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cost", totalCost))

	return nil
}

// completeSourceCardAction increments usage counts and consumes an action
// for the card action that triggered this card draw selection
func (a *ConfirmCardDrawAction) completeSourceCardAction(
	g *game.Game,
	p *player.Player,
	selection *player.PendingCardDrawSelection,
	log *zap.Logger,
) {
	// Increment usage counts for the source card action
	actions := p.Actions().List()
	for i := range actions {
		if actions[i].CardID == selection.SourceCardID && actions[i].BehaviorIndex == selection.SourceBehaviorIndex {
			actions[i].TimesUsedThisTurn++
			actions[i].TimesUsedThisGeneration++
			log.Debug("📊 Incremented action usage counts from card draw confirmation",
				zap.String("card_id", selection.SourceCardID),
				zap.Int("behavior_index", selection.SourceBehaviorIndex),
				zap.Int("times_used_this_turn", actions[i].TimesUsedThisTurn),
				zap.Int("times_used_this_generation", actions[i].TimesUsedThisGeneration))
			break
		}
	}
	p.Actions().SetActions(actions)

	// Consume the player action
	a.ConsumePlayerAction(g, log)
}
