package confirmation

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmSellPatentsAction handles the business logic for confirming sell patents card selection
// This is Phase 2: processes the selected cards and awards credits
type ConfirmSellPatentsAction struct {
	baseaction.BaseAction
}

// NewConfirmSellPatentsAction creates a new confirm sell patents action
func NewConfirmSellPatentsAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConfirmSellPatentsAction {
	return &ConfirmSellPatentsAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the confirm sell patents action (Phase 2: process card selection)
func (a *ConfirmSellPatentsAction) Execute(ctx context.Context, gameID string, playerID string, selectedCardIDs []string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_sell_patents"),
		zap.Int("cards_selected", len(selectedCardIDs)),
	)
	log.Info("🏛️ Confirming sell patents card selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	pendingCardSelection := player.Selection().GetPendingCardSelection()
	if pendingCardSelection == nil {
		log.Warn("No pending card selection found")
		return fmt.Errorf("no pending card selection found")
	}

	if pendingCardSelection.Source != "sell-patents" {
		log.Warn("Pending card selection is not for sell patents",
			zap.String("source", pendingCardSelection.Source))
		return fmt.Errorf("pending card selection is not for sell patents")
	}

	if len(selectedCardIDs) < pendingCardSelection.MinCards {
		log.Warn("Too few cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("min_required", pendingCardSelection.MinCards))
		return fmt.Errorf("must select at least %d cards", pendingCardSelection.MinCards)
	}

	if len(selectedCardIDs) > pendingCardSelection.MaxCards {
		log.Warn("Too many cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("max_allowed", pendingCardSelection.MaxCards))
		return fmt.Errorf("cannot select more than %d cards", pendingCardSelection.MaxCards)
	}

	availableCardsMap := make(map[string]bool)
	for _, cardID := range pendingCardSelection.AvailableCards {
		availableCardsMap[cardID] = true
	}

	for _, cardID := range selectedCardIDs {
		if !availableCardsMap[cardID] {
			log.Warn("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s is not available for selection", cardID)
		}
	}

	totalReward := 0
	for _, cardID := range selectedCardIDs {
		totalReward += pendingCardSelection.CardRewards[cardID]
	}

	if totalReward > 0 {
		player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: totalReward,
		})

		resources := player.Resources().Get()
		log.Info("💰 Awarded credits for sold cards",
			zap.Int("cards_sold", len(selectedCardIDs)),
			zap.Int("credits_earned", totalReward),
			zap.Int("new_credits", resources.Credits))
	}

	for _, cardID := range selectedCardIDs {
		removed := player.Hand().RemoveCard(cardID)
		if !removed {
			log.Warn("Failed to remove card from hand", zap.String("card_id", cardID))
		}
	}

	if err := g.Deck().Discard(ctx, selectedCardIDs); err != nil {
		log.Error("Failed to discard sold cards to discard pile", zap.Error(err))
		return fmt.Errorf("failed to discard sold cards: %w", err)
	}

	log.Info("🗑️ Sold cards added to discard pile", zap.Int("cards_removed", len(selectedCardIDs)))

	player.Selection().SetPendingCardSelection(nil)

	if len(selectedCardIDs) > 0 {
		a.ConsumePlayerAction(g, log)

		cardsSold := len(selectedCardIDs)
		creditOutputs := []game.CalculatedOutput{
			{ResourceType: string(shared.ResourceCredit), Amount: totalReward, IsScaled: false},
		}

		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:          "Sell Patents",
			PlayerID:          playerID,
			SourceType:        game.SourceTypeStandardProject,
			CalculatedOutputs: creditOutputs,
		})

		displayData := &game.LogDisplayData{
			Behaviors: []shared.CardBehavior{{
				Outputs: []shared.ResourceCondition{{
					ResourceType: shared.ResourceCardDraw,
					Amount:       cardsSold,
					Target:       "self-player",
				}},
			}},
		}
		a.WriteStateLogFull(ctx, g, "Standard Project: Sell Patents", game.SourceTypeStandardProject, playerID, "Sold patents", nil, creditOutputs, displayData)
	}

	log.Info("✅ Sell patents completed successfully",
		zap.Int("cards_sold", len(selectedCardIDs)),
		zap.Int("credits_earned", totalReward))
	return nil
}
