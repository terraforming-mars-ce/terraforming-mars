package confirmation

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"
	gameaction "terraforming-mars-backend/internal/action/game"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmProductionCardsAction handles the business logic for confirming production card selection
type ConfirmProductionCardsAction struct {
	baseaction.BaseAction
	finalScoringAction *gameaction.FinalScoringAction
}

// NewConfirmProductionCardsAction creates a new confirm production cards action
func NewConfirmProductionCardsAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	finalScoringAction *gameaction.FinalScoringAction,
	logger *zap.Logger,
) *ConfirmProductionCardsAction {
	return &ConfirmProductionCardsAction{
		BaseAction:         baseaction.NewBaseAction(gameRepo, cardRegistry),
		finalScoringAction: finalScoringAction,
	}
}

// Execute performs the confirm production cards action
func (a *ConfirmProductionCardsAction) Execute(ctx context.Context, gameID string, playerID string, selectedCardIDs []string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_production_cards"),
		zap.Strings("selected_card_ids", selectedCardIDs),
	)
	log.Debug("Player confirming production card selection")

	g, err := a.GameRepository().Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.CurrentPhase() != shared.GamePhaseProductionAndCardDraw {
		log.Warn("Game is not in production phase",
			zap.String("current_phase", string(g.CurrentPhase())),
			zap.String("expected_phase", string(shared.GamePhaseProductionAndCardDraw)))
		return fmt.Errorf("game is not in production phase")
	}

	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	productionPhase := g.GetProductionPhase(playerID)
	if productionPhase == nil {
		log.Error("Player not in production phase")
		return fmt.Errorf("player not in production phase")
	}

	if productionPhase.SelectionComplete {
		log.Error("Production selection already complete")
		return fmt.Errorf("production selection already complete")
	}

	availableSet := make(map[string]bool)
	for _, id := range productionPhase.AvailableCards {
		availableSet[id] = true
	}

	for _, cardID := range selectedCardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	calc := gamecards.NewRequirementModifierCalculator(a.CardRegistry())
	cardBuyDiscounts := calc.CalculateActionDiscounts(player, shared.ActionCardBuying)
	costPerCard := max(3-cardBuyDiscounts[shared.ResourceCredit], 0)
	cost := len(selectedCardIDs) * costPerCard

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

	resources = player.Resources().Get()
	log.Debug("Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", resources.Credits))

	log.Debug("Adding cards to player hand",
		zap.Strings("card_ids", selectedCardIDs),
		zap.Int("count", len(selectedCardIDs)))

	for _, cardID := range selectedCardIDs {
		player.Hand().AddCard(cardID)
	}

	log.Debug("Cards added to hand",
		zap.Strings("card_ids_added", selectedCardIDs),
		zap.Int("card_count", len(selectedCardIDs)))

	productionPhase.SelectionComplete = true
	if err := g.SetProductionPhase(ctx, playerID, productionPhase); err != nil {
		log.Error("Failed to update production phase", zap.Error(err))
		return fmt.Errorf("failed to update production phase: %w", err)
	}

	log.Debug("Production selection marked complete")

	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, p := range allPlayers {
		if p.HasExited() {
			continue
		}
		pPhase := g.GetProductionPhase(p.ID())
		if pPhase == nil || !pPhase.SelectionComplete {
			allComplete = false
			break
		}
	}

	if allComplete {
		for _, p := range allPlayers {
			if err := g.SetProductionPhase(ctx, p.ID(), nil); err != nil {
				log.Warn("Failed to clear production phase",
					zap.String("player_id", p.ID()),
					zap.Error(err))
			}
		}

		if g.GlobalParameters().IsMaxed() {
			log.Debug("All global parameters maxed after final production - triggering final scoring")
			if err := a.finalScoringAction.Execute(ctx, gameID); err != nil {
				log.Error("Failed to execute final scoring", zap.Error(err))
				return fmt.Errorf("failed to execute final scoring: %w", err)
			}
			log.Info("Game ended, final scores calculated")
			return nil
		}

		log.Debug("All players completed production selection, advancing to action phase")

		if err := g.UpdatePhase(ctx, shared.GamePhaseAction); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			return fmt.Errorf("failed to transition game phase: %w", err)
		}

		turnOrder := g.TurnOrder()
		activeCount := 0
		firstPlayerID := ""
		for _, id := range turnOrder {
			p, _ := g.GetPlayer(id)
			if p != nil && !p.HasExited() {
				activeCount++
				if firstPlayerID == "" {
					firstPlayerID = id
				}
			}
		}
		if firstPlayerID != "" {
			availableActions := 2
			if activeCount == 1 {
				availableActions = -1
			}
			if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}
			log.Debug("Set first player turn with actions",
				zap.String("player_id", firstPlayerID),
				zap.Int("available_actions", availableActions))
		}

		for _, p := range allPlayers {
			if !p.HasExited() {
				p.Actions().ResetGenerationCounts()
			}
		}
	}

	log.Info("Production cards selected")
	return nil
}
