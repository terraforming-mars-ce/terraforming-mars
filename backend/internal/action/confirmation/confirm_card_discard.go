package confirmation

import (
	"context"
	"fmt"
	"slices"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ConfirmCardDiscardAction handles the business logic for confirming card discard selection
type ConfirmCardDiscardAction struct {
	baseaction.BaseAction
}

// NewConfirmCardDiscardAction creates a new confirm card discard action
func NewConfirmCardDiscardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmCardDiscardAction {
	return &ConfirmCardDiscardAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute performs the confirm card discard action
// cardsToDiscard: card IDs from hand to discard (empty = skip if optional)
func (a *ConfirmCardDiscardAction) Execute(ctx context.Context, gameID string, playerID string, cardsToDiscard []string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_card_discard"),
		zap.Int("cards_to_discard", len(cardsToDiscard)),
	)
	log.Debug("Confirming card discard selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	selection := p.Selection().GetPendingCardDiscardSelection()
	if selection == nil {
		log.Warn("No pending card discard selection found")
		return fmt.Errorf("no pending card discard selection found")
	}

	// Validate discard count
	if len(cardsToDiscard) < selection.MinCards {
		log.Warn("Not enough cards to discard",
			zap.Int("selected", len(cardsToDiscard)),
			zap.Int("min_required", selection.MinCards))
		return fmt.Errorf("must discard at least %d card(s), selected %d", selection.MinCards, len(cardsToDiscard))
	}

	if len(cardsToDiscard) > selection.MaxCards {
		log.Warn("Too many cards to discard",
			zap.Int("selected", len(cardsToDiscard)),
			zap.Int("max_allowed", selection.MaxCards))
		return fmt.Errorf("can discard at most %d card(s), selected %d", selection.MaxCards, len(cardsToDiscard))
	}

	// Validate all cards are in hand
	handCards := p.Hand().Cards()
	for _, cardID := range cardsToDiscard {
		if !slices.Contains(handCards, cardID) {
			log.Warn("Card not in hand", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in player's hand", cardID)
		}
	}

	// Remove discarded cards from hand
	for _, cardID := range cardsToDiscard {
		p.Hand().RemoveCard(cardID)
	}

	if len(cardsToDiscard) > 0 {
		if err := g.Deck().Discard(ctx, cardsToDiscard); err != nil {
			log.Error("Failed to discard cards to discard pile", zap.Error(err))
			return fmt.Errorf("failed to discard cards: %w", err)
		}
		log.Debug("Discarded cards from hand to discard pile",
			zap.Int("count", len(cardsToDiscard)),
			zap.Strings("card_ids", cardsToDiscard))
	}

	// Apply pending outputs if player actually discarded (or if discard was mandatory with min=0)
	if len(cardsToDiscard) > 0 && len(selection.PendingOutputs) > 0 {
		selfOutputs, err := a.applyPendingOutputs(ctx, g, p, selection, log)
		if err != nil {
			log.Error("Failed to apply pending outputs after discard", zap.Error(err))
			return fmt.Errorf("failed to apply pending outputs: %w", err)
		}

		// Add triggered effect for self-player: discard + draws
		calculatedOutputs := []game.CalculatedOutput{
			{ResourceType: string(shared.ResourceCardDiscard), Amount: len(cardsToDiscard)},
		}
		calculatedOutputs = append(calculatedOutputs, selfOutputs...)
		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:          selection.Source,
			PlayerID:          p.ID(),
			SourceType:        game.SourceTypeCardPlay,
			CalculatedOutputs: calculatedOutputs,
		})
	}

	// Clear the pending selection
	p.Selection().SetPendingCardDiscardSelection(nil)

	log.Info("Card discard confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("cards_discarded", len(cardsToDiscard)))

	return nil
}

// applyPendingOutputs applies the outputs after a successful discard.
// Returns calculated outputs for the self-player (for triggered effect notifications).
func (a *ConfirmCardDiscardAction) applyPendingOutputs(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	selection *player.PendingCardDiscardSelection,
	log *zap.Logger,
) ([]game.CalculatedOutput, error) {
	var selfOutputs []game.CalculatedOutput

	for _, output := range selection.PendingOutputs {
		if output.ResourceType == shared.ResourceCardDraw {
			if output.Target == "all-opponents" {
				for _, opponent := range g.GetAllPlayers() {
					if opponent.ID() == p.ID() {
						continue
					}
					drawnCards, err := g.Deck().DrawProjectCards(ctx, output.Amount)
					if err != nil {
						log.Warn("Failed to draw cards for opponent",
							zap.String("opponent_id", opponent.ID()),
							zap.Error(err))
						continue
					}
					baseaction.AddCardsToPlayerHand(drawnCards, opponent, g, a.CardRegistry(), log)
					log.Debug("Opponent drew cards",
						zap.String("opponent_id", opponent.ID()),
						zap.Int("count", len(drawnCards)))

					g.AddTriggeredEffect(game.TriggeredEffect{
						CardName:   selection.Source,
						PlayerID:   opponent.ID(),
						SourceType: game.SourceTypeCardPlay,
						CalculatedOutputs: []game.CalculatedOutput{
							{ResourceType: string(shared.ResourceCardDraw), Amount: len(drawnCards)},
						},
					})
				}
				continue
			}

			drawnCards, err := g.Deck().DrawProjectCards(ctx, output.Amount)
			if err != nil {
				return nil, fmt.Errorf("failed to draw cards: %w", err)
			}
			baseaction.AddCardsToPlayerHand(drawnCards, p, g, a.CardRegistry(), log)
			selfOutputs = append(selfOutputs, game.CalculatedOutput{
				ResourceType: string(shared.ResourceCardDraw),
				Amount:       len(drawnCards),
			})
			log.Debug("Drew cards after discard",
				zap.Int("count", len(drawnCards)))
			continue
		}

		// For non-card-draw outputs, use the behavior applier
		applier := gamecards.NewBehaviorApplier(p, g, selection.Source, log).
			WithSourceCardID(selection.SourceCardID).
			WithCardRegistry(a.CardRegistry()).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, a.CardRegistry()))
		if err := applier.ApplyOutputs(ctx, []shared.ResourceCondition{output}); err != nil {
			return nil, err
		}
	}
	return selfOutputs, nil
}
