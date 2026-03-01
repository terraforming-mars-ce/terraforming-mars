package action

import (
	"context"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// SubscribePassiveEffectToEvents subscribes passive effects to relevant domain events
// This function is called when cards with passive effects are played or corporations are selected
func SubscribePassiveEffectToEvents(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	log *zap.Logger,
	cardRegistry ...cards.CardRegistry,
) {
	var cr cards.CardRegistry
	if len(cardRegistry) > 0 {
		cr = cardRegistry[0]
	}
	for _, trigger := range effect.Behavior.Triggers {
		// Only handle auto triggers with conditions (passive effects)
		if trigger.Type != "auto" || trigger.Condition == nil {
			continue
		}

		var subID events.SubscriptionID

		// Handle placement-bonus-gained trigger
		if trigger.Condition.Type == "placement-bonus-gained" {
			subID = subscribePlacementBonusEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle city-placed trigger
		if trigger.Condition.Type == "city-placed" {
			subID = subscribeCityPlacedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle ocean-placed trigger
		if trigger.Condition.Type == "ocean-placed" {
			subID = subscribeOceanPlacedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle tag-played trigger
		if trigger.Condition.Type == "tag-played" {
			subID = subscribeTagPlayedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle card-played trigger with selectors
		if trigger.Condition.Type == "card-played" {
			subID = subscribeCardPlayedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle standard-project-played trigger
		if trigger.Condition.Type == "standard-project-played" {
			subID = subscribeStandardProjectPlayedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle tile-placed trigger (production based on placement bonus type)
		if trigger.Condition.Type == "tile-placed" {
			subID = subscribeTilePlacedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Register subscription for cleanup when effect is removed
		if subID != "" {
			p.Effects().RegisterSubscription(effect.CardID, subID)
		}
	}
}

// subscribePlacementBonusEffect subscribes to PlacementBonusGainedEvent
func subscribePlacementBonusEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.PlacementBonusGainedEvent) {
		// Only process if event is for this game and player
		if event.GameID != g.ID() {
			return
		}

		// Check target condition (self-player, any-player, etc.)
		target := "self-player" // Default
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return // Effect only applies to self
		}

		// Check if selectors match the bonus resources
		if len(trigger.Condition.Selectors) > 0 {
			matchFound := false
			for _, sel := range trigger.Condition.Selectors {
				for _, resource := range sel.Resources {
					if _, exists := event.Resources[resource]; exists {
						matchFound = true
						break
					}
				}
				if matchFound {
					break
				}
			}
			if !matchFound {
				return // No matching resources in the bonus
			}
		}

		// Condition matched! Apply the effect outputs using BehaviorApplier
		log.Info("🎴 Passive effect triggered",
			zap.String("card_name", effect.CardName),
			zap.String("trigger_type", trigger.Condition.Type),
			zap.Any("resources_gained", event.Resources))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to PlacementBonusGainedEvent",
		zap.String("card_name", effect.CardName))

	return subID
}

// subscribeCityPlacedEffect subscribes to TilePlacedEvent for city placements
func subscribeCityPlacedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.TilePlacedEvent) {
		// Only process if event is for this game
		if event.GameID != g.ID() {
			return
		}

		// Only process city tile placements
		// TileType is ResourceCityTile constant value: "city-tile"
		if event.TileType != string(shared.ResourceCityTile) {
			return
		}

		// Check target condition (self-player, any-player, etc.)
		target := "self-player" // Default
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return // Effect only applies to self
		}

		// Check location condition
		location := "anywhere" // Default
		if trigger.Condition.Location != nil {
			location = *trigger.Condition.Location
		}

		// For now, we treat all tile placements as "mars" or "anywhere"
		// Future: implement Phobos/colony distinction if needed
		if location != "anywhere" && location != "mars" {
			return // Location doesn't match
		}

		// Condition matched! Apply the effect outputs using BehaviorApplier
		log.Info("🎴 Passive effect triggered (city placement)",
			zap.String("card_name", effect.CardName),
			zap.String("player_id", p.ID()),
			zap.String("placed_by", event.PlayerID),
			zap.String("tile_type", event.TileType))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to TilePlacedEvent (city)",
		zap.String("card_name", effect.CardName))

	return subID
}

// subscribeOceanPlacedEffect subscribes to TilePlacedEvent for ocean placements
func subscribeOceanPlacedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr gamecards.CardRegistryInterface,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.TilePlacedEvent) {
		if event.GameID != g.ID() {
			return
		}

		if event.TileType != string(shared.ResourceOceanTile) {
			return
		}

		target := "self-player"
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return
		}

		location := "anywhere"
		if trigger.Condition.Location != nil {
			location = *trigger.Condition.Location
		}

		if location != "anywhere" && location != "mars" {
			return
		}

		log.Info("🎴 Passive effect triggered (ocean placement)",
			zap.String("card_name", effect.CardName),
			zap.String("player_id", p.ID()),
			zap.String("placed_by", event.PlayerID),
			zap.String("tile_type", event.TileType))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect)
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to TilePlacedEvent (ocean)",
		zap.String("card_name", effect.CardName))

	return subID
}

func subscribeTagPlayedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.TagPlayedEvent) {
		if event.GameID != g.ID() {
			return
		}

		target := "self-player"
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return
		}

		// Check if selectors match the played tag
		if len(trigger.Condition.Selectors) > 0 {
			matchFound := false
			for _, sel := range trigger.Condition.Selectors {
				for _, tag := range sel.Tags {
					if string(tag) == event.Tag {
						matchFound = true
						break
					}
				}
				if matchFound {
					break
				}
			}
			if !matchFound {
				return
			}
		}

		log.Info("🎴 Passive effect triggered (tag played)",
			zap.String("card_name", effect.CardName),
			zap.String("effect_owner", p.ID()),
			zap.String("tag_played_by", event.PlayerID),
			zap.String("tag", event.Tag))

		// Check if this effect requires card-discard input (e.g., Mars University)
		if gamecards.HasCardDiscardInput(effect.Behavior) {
			createPassiveCardDiscard(p, effect, log)
			return
		}

		// Check if this effect has choices requiring player selection (e.g., Olympus Conference, Viral Enhancers)
		if gamecards.HasChoices(effect.Behavior) {
			createPassiveBehaviorChoice(p, effect, log)
			return
		}

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to TagPlayedEvent",
		zap.String("card_name", effect.CardName))

	return subID
}

func subscribeCardPlayedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.CardPlayedEvent) {
		if event.GameID != g.ID() {
			return
		}

		target := "self-player"
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return
		}

		if cr == nil {
			return
		}

		card, err := cr.GetByID(event.CardID)
		if err != nil {
			return
		}

		// Check if the card matches any selector
		if len(trigger.Condition.Selectors) > 0 {
			if !gamecards.MatchesAnySelector(card, trigger.Condition.Selectors) {
				return
			}
		}

		log.Info("🎴 Passive effect triggered (card played)",
			zap.String("card_name", effect.CardName),
			zap.String("effect_owner", p.ID()),
			zap.String("card_played_by", event.PlayerID),
			zap.String("card_played", event.CardName))

		if gamecards.HasChoices(effect.Behavior) {
			createPassiveBehaviorChoice(p, effect, log)
			return
		}

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to CardPlayedEvent",
		zap.String("card_name", effect.CardName))

	return subID
}

func subscribeStandardProjectPlayedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.StandardProjectPlayedEvent) {
		if event.GameID != g.ID() {
			return
		}

		target := "self-player"
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return
		}

		// Check selectors for cost and project type matching
		if len(trigger.Condition.Selectors) > 0 {
			matched := false
			for _, sel := range trigger.Condition.Selectors {
				// Check cost requirement
				if sel.RequiredOriginalCost != nil {
					if sel.RequiredOriginalCost.Min != nil && event.ProjectCost < *sel.RequiredOriginalCost.Min {
						continue
					}
					if sel.RequiredOriginalCost.Max != nil && event.ProjectCost > *sel.RequiredOriginalCost.Max {
						continue
					}
				}
				// Check project type match (if specified)
				if len(sel.StandardProjects) > 0 {
					if !gamecards.MatchesStandardProjectSelector(shared.StandardProject(event.ProjectType), sel) {
						continue
					}
				}
				matched = true
				break
			}
			if !matched {
				return
			}
		}

		log.Info("🎴 Passive effect triggered (standard project played)",
			zap.String("card_name", effect.CardName),
			zap.String("effect_owner", p.ID()),
			zap.String("project_type", event.ProjectType),
			zap.Int("project_cost", event.ProjectCost))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to StandardProjectPlayedEvent",
		zap.String("card_name", effect.CardName))

	return subID
}

func subscribeTilePlacedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.PlacementBonusGainedEvent) {
		if event.GameID != g.ID() {
			return
		}

		target := "self-player"
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-card" {
			if event.SourceCardID != effect.CardID {
				return
			}
		} else if target == "self-player" && event.PlayerID != p.ID() {
			return
		}

		if len(trigger.Condition.OnBonusType) > 0 {
			matchFound := false
			for _, requiredBonus := range trigger.Condition.OnBonusType {
				if _, exists := event.Resources[requiredBonus]; exists {
					matchFound = true
					break
				}
			}
			if !matchFound {
				return
			}
		}

		log.Info("🎴 Passive effect triggered (tile placed on bonus)",
			zap.String("card_name", effect.CardName),
			zap.String("trigger_type", trigger.Condition.Type),
			zap.Any("bonus_resources", event.Resources))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to PlacementBonusGainedEvent (tile-placed)",
		zap.String("card_name", effect.CardName))

	return subID
}

// createPassiveCardDiscard creates a pending card discard selection from a passive effect
// Used for effects like Mars University that require player to optionally discard before gaining outputs
func createPassiveCardDiscard(p *player.Player, effect player.CardEffect, log *zap.Logger) {
	// Find card-discard inputs to determine min/max
	minCards := 0
	maxCards := 0
	for _, input := range effect.Behavior.Inputs {
		if input.ResourceType == shared.ResourceCardDiscard {
			if !input.Optional {
				minCards = input.Amount
			}
			maxCards = input.Amount
			break
		}
	}

	// Skip if player has no cards to discard
	if len(p.Hand().Cards()) == 0 {
		log.Info("🗑️ Skipping card discard - player has no cards in hand",
			zap.String("card_name", effect.CardName))
		return
	}

	p.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards:       minCards,
		MaxCards:       maxCards,
		Source:         effect.CardName,
		SourceCardID:   effect.CardID,
		PendingOutputs: effect.Behavior.Outputs,
	})

	log.Info("🗑️ Created pending card discard selection from passive effect",
		zap.String("card_name", effect.CardName),
		zap.Int("min_cards", minCards),
		zap.Int("max_cards", maxCards))
}

// createPassiveBehaviorChoice creates a pending behavior choice selection from a passive effect
// Used for effects like Viral Enhancers and Olympus Conference that require player to choose between options
func createPassiveBehaviorChoice(p *player.Player, effect player.CardEffect, log *zap.Logger) {
	p.Selection().SetPendingBehaviorChoiceSelection(&player.PendingBehaviorChoiceSelection{
		Choices:      effect.Behavior.Choices,
		Source:       effect.CardName,
		SourceCardID: effect.CardID,
	})

	log.Info("🔀 Created pending behavior choice selection from passive effect",
		zap.String("card_name", effect.CardName),
		zap.Int("num_choices", len(effect.Behavior.Choices)))
}
