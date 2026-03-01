package turn_management

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectPreludeCardsAction handles the business logic for selecting prelude cards
type SelectPreludeCardsAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewSelectPreludeCardsAction creates a new select prelude cards action
func NewSelectPreludeCardsAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectPreludeCardsAction {
	return &SelectPreludeCardsAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the select prelude cards action
func (a *SelectPreludeCardsAction) Execute(ctx context.Context, gameID string, playerID string, preludeIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_prelude_cards"),
		zap.Strings("prelude_ids", preludeIDs),
	)
	log.Info("🃏 Player selecting prelude cards")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.CurrentPhase() != game.GamePhasePreludeSelection {
		log.Error("Game not in prelude selection phase", zap.String("phase", string(g.CurrentPhase())))
		return fmt.Errorf("game not in prelude selection phase")
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	preludePhase := g.GetSelectPreludeCardsPhase(playerID)
	if preludePhase == nil {
		log.Error("Player not in prelude selection phase")
		return fmt.Errorf("not in prelude selection phase")
	}

	if len(preludeIDs) != preludePhase.MaxSelectable {
		log.Error("Wrong number of preludes selected",
			zap.Int("selected", len(preludeIDs)),
			zap.Int("required", preludePhase.MaxSelectable))
		return fmt.Errorf("must select exactly %d preludes, got %d", preludePhase.MaxSelectable, len(preludeIDs))
	}

	availableSet := make(map[string]bool, len(preludePhase.AvailablePreludes))
	for _, id := range preludePhase.AvailablePreludes {
		availableSet[id] = true
	}
	for _, id := range preludeIDs {
		if !availableSet[id] {
			log.Error("Selected prelude not available", zap.String("prelude_id", id))
			return fmt.Errorf("prelude %s not available for selection", id)
		}
	}

	for _, preludeID := range preludeIDs {
		if err := a.applyPrelude(ctx, g, p, preludeID, log); err != nil {
			return fmt.Errorf("failed to apply prelude %s: %w", preludeID, err)
		}
	}

	selectedSet := make(map[string]bool, len(preludeIDs))
	for _, id := range preludeIDs {
		selectedSet[id] = true
	}
	var unselected []string
	for _, id := range preludePhase.AvailablePreludes {
		if !selectedSet[id] {
			unselected = append(unselected, id)
		}
	}
	if len(unselected) > 0 {
		if err := g.Deck().Discard(ctx, unselected); err != nil {
			log.Error("Failed to discard unselected preludes", zap.Error(err))
			return fmt.Errorf("failed to discard unselected preludes: %w", err)
		}
		log.Info("🗑️ Unselected preludes discarded", zap.Int("count", len(unselected)))
	}

	if err := g.SetSelectPreludeCardsPhase(ctx, playerID, nil); err != nil {
		log.Error("Failed to clear prelude phase", zap.Error(err))
		return fmt.Errorf("failed to clear prelude phase: %w", err)
	}
	log.Info("✅ Prelude selection marked complete")

	if err := a.checkAndAdvancePhase(ctx, g, log); err != nil {
		return err
	}

	log.Info("🎉 Prelude card selection completed successfully")
	return nil
}

// applyPrelude applies a single prelude card's effects to the player
func (a *SelectPreludeCardsAction) applyPrelude(ctx context.Context, g *game.Game, p *player.Player, preludeID string, log *zap.Logger) error {
	card, err := a.cardRegistry.GetByID(preludeID)
	if err != nil {
		log.Error("Failed to fetch prelude card", zap.String("prelude_id", preludeID), zap.Error(err))
		return fmt.Errorf("prelude card not found: %s", preludeID)
	}

	if card.Type != gamecards.CardTypePrelude {
		log.Error("Card is not a prelude", zap.String("card_type", string(card.Type)))
		return fmt.Errorf("card %s is not a prelude card", preludeID)
	}

	tags := make([]string, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = string(tag)
	}
	p.PlayedCards().AddCard(card.ID, card.Name, string(card.Type), tags)
	log.Info("✅ Prelude added to played cards", zap.String("prelude_id", preludeID), zap.String("name", card.Name))

	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		_, outputs := behavior.ExtractInputsOutputs(nil)

		log.Info("✨ Applying prelude behavior",
			zap.String("prelude", card.Name),
			zap.Int("index", behaviorIndex),
			zap.Int("output_count", len(outputs)))

		applier := gamecards.NewBehaviorApplier(p, g, card.Name, log).
			WithSourceCardID(card.ID).
			WithCardRegistry(a.cardRegistry).
			WithSourceType(game.SourceTypeCardPlay).
			WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, a.cardRegistry))

		calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
		if err != nil {
			return fmt.Errorf("failed to apply prelude behavior %d: %w", behaviorIndex, err)
		}

		if len(calculatedOutputs) > 0 {
			g.AddTriggeredEffect(game.TriggeredEffect{
				CardName:          card.Name,
				PlayerID:          p.ID(),
				SourceType:        game.SourceTypeCardPlay,
				Behaviors:         []shared.CardBehavior{behavior},
				CalculatedOutputs: calculatedOutputs,
			})
		}

		if gamecards.HasPersistentEffects(behavior) {
			effect := player.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			p.Effects().AddEffect(effect)

			events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
				GameID:    g.ID(),
				PlayerID:  p.ID(),
				Timestamp: time.Now(),
			})
		}
	}

	// Register passive effects
	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasConditionalTrigger(behavior) {
			continue
		}

		effect := player.CardEffect{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: behaviorIndex,
			Behavior:      behavior,
		}
		p.Effects().AddEffect(effect)
		baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, a.cardRegistry)
	}

	// Register manual actions
	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasManualTrigger(behavior) {
			continue
		}

		p.Actions().AddAction(player.CardAction{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: behaviorIndex,
			Behavior:      behavior,
		})
	}

	return nil
}

// checkAndAdvancePhase checks if all players have completed prelude selection and advances to starting card selection
func (a *SelectPreludeCardsAction) checkAndAdvancePhase(ctx context.Context, g *game.Game, log *zap.Logger) error {
	allPlayers := g.GetAllPlayers()

	for _, p := range allPlayers {
		if g.GetSelectPreludeCardsPhase(p.ID()) != nil {
			log.Info("⏳ Waiting for other players to complete prelude selection")
			return nil
		}
		if g.GetPendingTileSelection(p.ID()) != nil {
			log.Info("⏳ Waiting for player to complete pending tile selection", zap.String("player_id", p.ID()))
			return nil
		}
		if g.GetPendingTileSelectionQueue(p.ID()) != nil {
			log.Info("⏳ Waiting for player to complete pending tile queue", zap.String("player_id", p.ID()))
			return nil
		}
	}

	log.Info("🎉 All players completed prelude selection, advancing to starting card selection")

	if err := g.UpdatePhase(ctx, game.GamePhaseStartingCardSelection); err != nil {
		log.Error("Failed to transition to starting card selection phase", zap.Error(err))
		return fmt.Errorf("failed to transition to starting card selection phase: %w", err)
	}

	if err := a.distributeProjectCards(ctx, g, allPlayers); err != nil {
		log.Error("Failed to distribute project cards", zap.Error(err))
		return fmt.Errorf("failed to distribute project cards: %w", err)
	}
	log.Info("✅ Project cards distributed to all players")

	return nil
}

func (a *SelectPreludeCardsAction) distributeProjectCards(ctx context.Context, g *game.Game, players []*player.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		projectCardIDs, err := deck.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID(), err)
		}

		selectionPhase := &player.SelectStartingCardsPhase{
			AvailableCards: projectCardIDs,
		}
		if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), selectionPhase); err != nil {
			return fmt.Errorf("failed to set selection phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}
