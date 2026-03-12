package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CorporationProcessor handles applying corporation card effects
type CorporationProcessor struct {
	cardRegistry CardRegistryInterface
	logger       *zap.Logger
}

// NewCorporationProcessor creates a new corporation processor
func NewCorporationProcessor(cardRegistry CardRegistryInterface, logger *zap.Logger) *CorporationProcessor {
	return &CorporationProcessor{
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// ApplyStartingEffects processes ONLY auto-corporation-start behaviors
// and applies starting resources/production
func (p *CorporationProcessor) ApplyStartingEffects(
	ctx context.Context,
	card *Card,
	pl *player.Player,
	g *game.Game,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", pl.ID()),
	)

	log.Debug("Applying corporation starting effects")

	applier := NewBehaviorApplier(pl, g, card.Name, p.logger).
		WithSourceCardID(card.ID).
		WithCardRegistry(p.cardRegistry)

	// Process ONLY behaviors with auto-corporation-start trigger
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(ResourceTriggerAutoCorporationStart) {
				log.Debug("Found auto-corporation-start behavior",
					zap.Int("outputs", len(behavior.Outputs)))

				if err := applier.ApplyOutputs(ctx, behavior.Outputs); err != nil {
					return fmt.Errorf("failed to apply starting effects: %w", err)
				}
			}
		}
	}

	log.Debug("Corporation starting effects applied")
	return nil
}

// ApplyAutoEffects processes auto triggers WITHOUT conditions
// (e.g., payment-substitute for Helion)
func (p *CorporationProcessor) ApplyAutoEffects(
	ctx context.Context,
	card *Card,
	pl *player.Player,
	g *game.Game,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", pl.ID()),
	)

	log.Debug("Applying corporation auto effects")

	applier := NewBehaviorApplier(pl, g, card.Name, p.logger).
		WithSourceCardID(card.ID).
		WithCardRegistry(p.cardRegistry)

	// Process behaviors with auto trigger WITHOUT conditions
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			// Handle auto trigger WITHOUT conditions (immediate effects like payment-substitute)
			// Auto triggers WITH conditions are passive effects handled separately
			if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition == nil {
				log.Debug("Found auto behavior (no condition)",
					zap.Int("outputs", len(behavior.Outputs)))

				if err := applier.ApplyOutputs(ctx, behavior.Outputs); err != nil {
					return fmt.Errorf("failed to apply auto effects: %w", err)
				}
			}
		}
	}

	log.Debug("Corporation auto effects applied")
	return nil
}

// SetupForcedFirstAction processes auto-corporation-first-action behaviors and sets forced actions
func (p *CorporationProcessor) SetupForcedFirstAction(
	ctx context.Context,
	card *Card,
	g *game.Game,
	playerID string,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", playerID),
	)

	log.Debug("Checking for forced first action")

	// Process behaviors with auto-corporation-first-action trigger
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(ResourceTriggerAutoCorporationFirstAction) {
				log.Debug("Found auto-corporation-first-action behavior",
					zap.Int("outputs", len(behavior.Outputs)))

				// Check if this behavior has card-peek/card-take outputs (e.g. Valley Trust)
				if p.hasCardDrawOutputs(behavior) {
					if err := p.applyCardDrawForcedAction(ctx, behavior, card, g, playerID, log); err != nil {
						return fmt.Errorf("failed to apply card draw forced action: %w", err)
					}
					continue
				}

				// Create forced action based on individual outputs
				for _, output := range behavior.Outputs {
					if err := p.createForcedAction(ctx, output, card, g, playerID, log); err != nil {
						return fmt.Errorf("failed to create forced action: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// GetAutoEffects returns all auto effects (without conditions) from a corporation card
// These are behaviors with auto triggers without conditions (e.g., payment-substitute for Helion)
// They are applied immediately AND registered in effects list for display purposes
// This is a READ-ONLY helper that parses the card behaviors and returns CardEffect structs
// The action layer is responsible for adding these effects to the player
func (p *CorporationProcessor) GetAutoEffects(card *Card) []shared.CardEffect {
	var effects []shared.CardEffect

	// Iterate through all behaviors and find auto triggers without conditions
	for behaviorIndex, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			// Auto triggers WITHOUT conditions are immediate/permanent effects
			if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition == nil {
				effect := shared.CardEffect{
					CardID:        card.ID,
					CardName:      card.Name,
					BehaviorIndex: behaviorIndex,
					Behavior:      behavior,
				}
				effects = append(effects, effect)
			}
		}
	}

	return effects
}

// GetTriggerEffects returns all trigger effects (conditional triggers) from a corporation card
// These are behaviors with auto triggers that have conditions, for event subscription
// This is a READ-ONLY helper that parses the card behaviors and returns CardEffect structs
// The action layer is responsible for adding these effects to the player
func (p *CorporationProcessor) GetTriggerEffects(card *Card) []shared.CardEffect {
	var effects []shared.CardEffect

	// Iterate through all behaviors and find conditional triggers
	for behaviorIndex, behavior := range card.Behaviors {
		if HasConditionalTrigger(behavior) {
			effect := shared.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			effects = append(effects, effect)
		}
	}

	return effects
}

// GetManualActions returns all manual actions (manual triggers) from a corporation card
// This is a READ-ONLY helper that parses the card behaviors and returns CardAction structs
// The action layer is responsible for adding these actions to the player
func (p *CorporationProcessor) GetManualActions(card *Card) []shared.CardAction {
	var actions []shared.CardAction

	// Iterate through all behaviors and find manual triggers
	for behaviorIndex, behavior := range card.Behaviors {
		if HasManualTrigger(behavior) {
			action := shared.CardAction{
				CardID:                  card.ID,
				CardName:                card.Name,
				BehaviorIndex:           behaviorIndex,
				Behavior:                behavior,
				TimesUsedThisTurn:       0,
				TimesUsedThisGeneration: 0,
			}
			actions = append(actions, action)
		}
	}

	return actions
}

// hasCardDrawOutputs returns true if the behavior has card-peek or card-take outputs
func (p *CorporationProcessor) hasCardDrawOutputs(behavior shared.CardBehavior) bool {
	for _, output := range behavior.Outputs {
		switch output.ResourceType {
		case shared.ResourceCardPeek, shared.ResourceCardTake, shared.ResourceCardBuy:
			return true
		}
	}
	return false
}

// applyCardDrawForcedAction handles first-action behaviors with card-peek/card-take outputs.
// These need to be processed as a batch via ApplyCardDrawOutputs.
func (p *CorporationProcessor) applyCardDrawForcedAction(
	ctx context.Context,
	behavior shared.CardBehavior,
	card *Card,
	g *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	pl, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for card draw: %w", err)
	}

	applier := NewBehaviorApplier(pl, g, card.Name, log).
		WithSourceCardID(card.ID).
		WithCardRegistry(p.cardRegistry)

	_, err = applier.ApplyCardDrawOutputs(ctx, behavior.Outputs)
	if err != nil {
		return fmt.Errorf("failed to apply card draw outputs: %w", err)
	}

	action := &shared.ForcedFirstAction{
		ActionType:    "card-draw-selection",
		CorporationID: card.ID,
		Source:        "corporation-starting-action",
		Completed:     false,
		Description:   fmt.Sprintf("Draw and select cards (%s starting action)", card.Name),
	}
	if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
		return fmt.Errorf("failed to set forced card draw action: %w", err)
	}

	log.Debug("Set forced card draw selection action",
		zap.String("description", action.Description))

	return nil
}

// createForcedAction creates a forced first action based on the output.
// During starting_selection, only stores the ForcedFirstAction metadata without creating tile queues.
// Tile queues are created when transitioning to action phase to avoid conflicts with prelude tile placements.
func (p *CorporationProcessor) createForcedAction(
	ctx context.Context,
	output shared.ResourceCondition,
	card *Card,
	g *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	inStartingSelection := g.CurrentPhase() == shared.GamePhaseStartingSelection

	switch output.ResourceType {
	case shared.ResourceCityPlacement:
		action := &shared.ForcedFirstAction{
			ActionType:    "city-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place a city tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced city placement action: %w", err)
		}
		log.Debug("Set forced city placement action",
			zap.String("description", action.Description))

		if !inStartingSelection {
			queue := &shared.PendingTileSelectionQueue{
				Items:  []string{"city"},
				Source: "corporation-starting-action",
			}
			if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
				return fmt.Errorf("failed to queue tile placement: %w", err)
			}
			log.Debug("Queued city tile for placement")
		} else {
			log.Debug("Deferred city tile queue to action phase")
		}

		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	case shared.ResourceGreeneryPlacement:
		action := &shared.ForcedFirstAction{
			ActionType:    "greenery-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place a greenery tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced greenery placement action: %w", err)
		}
		log.Debug("Set forced greenery placement action",
			zap.String("description", action.Description))

		if !inStartingSelection {
			queue := &shared.PendingTileSelectionQueue{
				Items:  []string{"greenery"},
				Source: "corporation-starting-action",
			}
			if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
				return fmt.Errorf("failed to queue tile placement: %w", err)
			}
			log.Debug("Queued greenery tile for placement")
		} else {
			log.Debug("Deferred greenery tile queue to action phase")
		}

		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	case shared.ResourceOceanPlacement:
		action := &shared.ForcedFirstAction{
			ActionType:    "ocean-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place an ocean tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced ocean placement action: %w", err)
		}
		log.Debug("Set forced ocean placement action",
			zap.String("description", action.Description))

		if !inStartingSelection {
			queue := &shared.PendingTileSelectionQueue{
				Items:  []string{"ocean"},
				Source: "corporation-starting-action",
			}
			if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
				return fmt.Errorf("failed to queue tile placement: %w", err)
			}
			log.Debug("Queued ocean tile for placement")
		} else {
			log.Debug("Deferred ocean tile queue to action phase")
		}

		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	case shared.ResourceCardDraw:
		pl, err := g.GetPlayer(playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for card draw: %w", err)
		}

		applier := NewBehaviorApplier(pl, g, card.Name, log).
			WithSourceCardID(card.ID).
			WithCardRegistry(p.cardRegistry)

		if err := applier.ApplyOutputs(ctx, []shared.ResourceCondition{output}); err != nil {
			return fmt.Errorf("failed to apply card-draw output: %w", err)
		}
		log.Debug("Applied card-draw forced action",
			zap.Int("amount", output.Amount))

	default:
		log.Warn("Unhandled forced action type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}

// subscribeForcedActionCompletion subscribes to TilePlacedEvent to handle forced action completion
// When the last tile in a forced action is placed, this consumes 1 player action and clears the forced action
func (p *CorporationProcessor) subscribeForcedActionCompletion(
	ctx context.Context,
	g *game.Game,
	playerID string,
	source string,
	log *zap.Logger,
) {
	eventBus := g.EventBus()
	if eventBus == nil {
		log.Warn("No event bus available, cannot subscribe to forced action completion")
		return
	}

	// Subscribe to TilePlacedEvent
	events.Subscribe(eventBus, func(event events.TilePlacedEvent) {
		// Only handle events for this player
		if event.PlayerID != playerID {
			return
		}

		log.Debug("Received TilePlacedEvent for forced action check",
			zap.String("player_id", event.PlayerID),
			zap.String("tile_type", event.TileType))

		// Check if there's a forced first action for this player
		forcedAction := g.GetForcedFirstAction(playerID)
		if forcedAction == nil {
			log.Debug("No forced first action, ignoring event")
			return
		}

		// Check if the queue is now empty (last tile was placed)
		queue := g.GetPendingTileSelectionQueue(playerID)
		if queue != nil && len(queue.Items) > 0 {
			log.Debug("Tile queue still has items, waiting for more tiles",
				zap.Int("remaining_tiles", len(queue.Items)))
			return
		}

		// Queue is empty - forced action is complete!
		// Note: Forced first actions are FREE - they don't consume player actions
		log.Debug("Forced first action completed (free action)",
			zap.String("action_type", forcedAction.ActionType),
			zap.String("corporation_id", forcedAction.CorporationID))

		// Clear forced first action
		if err := g.SetForcedFirstAction(ctx, playerID, nil); err != nil {
			log.Error("Failed to clear forced first action", zap.Error(err))
		}
	})

	log.Debug("Subscribed to TilePlacedEvent for forced action completion",
		zap.String("player_id", playerID),
		zap.String("source", source))
}
