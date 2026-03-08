package action

import (
	"context"
	"slices"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
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

		// Handle global-parameter-raised trigger
		if trigger.Condition.Type == "global-parameter-raised" {
			subID = subscribeGlobalParameterRaisedEffect(ctx, g, p, effect, trigger, log, cr)
		}

		// Handle production-increased trigger
		if trigger.Condition.Type == "production-increased" {
			subID = subscribeProductionIncreasedEffect(ctx, g, p, effect, trigger, log, cr)
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
		log.Debug("Passive effect triggered",
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

	log.Debug("Subscribed passive effect to PlacementBonusGainedEvent",
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
		if event.TileType != board.TileTypeCity {
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
		log.Debug("Passive effect triggered (city placement)",
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

	log.Debug("Subscribed passive effect to TilePlacedEvent (city)",
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

		if event.TileType != board.TileTypeOcean {
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

		log.Debug("Passive effect triggered (ocean placement)",
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

	log.Debug("Subscribed passive effect to TilePlacedEvent (ocean)",
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

		log.Debug("Passive effect triggered (tag played)",
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

	log.Debug("Subscribed passive effect to TagPlayedEvent",
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

		log.Debug("Passive effect triggered (card played)",
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

	log.Debug("Subscribed passive effect to CardPlayedEvent",
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

		log.Debug("Passive effect triggered (standard project played)",
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

	log.Debug("Subscribed passive effect to StandardProjectPlayedEvent",
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

		log.Debug("Passive effect triggered (tile placed on bonus)",
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

	log.Debug("Subscribed passive effect to PlacementBonusGainedEvent (tile-placed)",
		zap.String("card_name", effect.CardName))

	return subID
}

func subscribeGlobalParameterRaisedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	_ shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	globalParams := getGlobalParametersFromSelectors(effect.Behavior.Triggers)

	applyPerStep := func(steps int, paramName string) {
		if steps <= 0 {
			return
		}
		log.Debug("Passive effect triggered (global parameter raised)",
			zap.String("card_name", effect.CardName),
			zap.String("player_id", p.ID()),
			zap.String("parameter", paramName),
			zap.Int("steps", steps))

		for i := 0; i < steps; i++ {
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
		}
	}

	if slices.Contains(globalParams, "venus") {
		subID := events.Subscribe(g.EventBus(), func(event events.VenusChangedEvent) {
			if event.GameID != g.ID() {
				return
			}
			steps := (event.NewValue - event.OldValue) / 2
			applyPerStep(steps, "venus")
		})
		p.Effects().RegisterSubscription(effect.CardID, subID)
	}

	if slices.Contains(globalParams, "temperature") {
		subID := events.Subscribe(g.EventBus(), func(event events.TemperatureChangedEvent) {
			if event.GameID != g.ID() {
				return
			}
			steps := (event.NewValue - event.OldValue) / 2
			applyPerStep(steps, "temperature")
		})
		p.Effects().RegisterSubscription(effect.CardID, subID)
	}

	if slices.Contains(globalParams, "oxygen") {
		subID := events.Subscribe(g.EventBus(), func(event events.OxygenChangedEvent) {
			if event.GameID != g.ID() {
				return
			}
			steps := event.NewValue - event.OldValue
			applyPerStep(steps, "oxygen")
		})
		p.Effects().RegisterSubscription(effect.CardID, subID)
	}

	log.Debug("Subscribed passive effect to global parameter raised events",
		zap.String("card_name", effect.CardName),
		zap.Strings("parameters", globalParams))

	// Subscriptions are registered internally per parameter, return empty to avoid duplicate registration by caller
	return ""
}

func getGlobalParametersFromSelectors(triggers []shared.Trigger) []string {
	for _, trigger := range triggers {
		if trigger.Condition == nil {
			continue
		}
		for _, sel := range trigger.Condition.Selectors {
			if len(sel.GlobalParameters) > 0 {
				return sel.GlobalParameters
			}
		}
	}
	return nil
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
		log.Debug("Skipping card discard - player has no cards in hand",
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

	log.Debug("Created pending card discard selection from passive effect",
		zap.String("card_name", effect.CardName),
		zap.Int("min_cards", minCards),
		zap.Int("max_cards", maxCards))
}

// resourceNameToProductionType maps event resource names to production resource types
var resourceNameToProductionType = map[string]shared.ResourceType{
	"credits":  shared.ResourceCreditProduction,
	"steel":    shared.ResourceSteelProduction,
	"titanium": shared.ResourceTitaniumProduction,
	"plants":   shared.ResourcePlantProduction,
	"energy":   shared.ResourceEnergyProduction,
	"heat":     shared.ResourceHeatProduction,
}

// subscribeProductionIncreasedEffect subscribes to ProductionChangedEvent for production increase triggers
func subscribeProductionIncreasedEffect(
	_ context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
	cr cards.CardRegistry,
) events.SubscriptionID {
	subID := events.Subscribe(g.EventBus(), func(event events.ProductionChangedEvent) {
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

		increase := event.NewProduction - event.OldProduction
		if increase <= 0 {
			return
		}

		// Check if this production type matches the trigger's resource type filter
		if len(trigger.Condition.ResourceTypes) > 0 {
			productionType, exists := resourceNameToProductionType[event.ResourceType]
			if !exists {
				return
			}
			if !slices.Contains(trigger.Condition.ResourceTypes, productionType) {
				return
			}
		}

		// Scale outputs by the production increase amount
		scaledOutputs := make([]shared.ResourceCondition, len(effect.Behavior.Outputs))
		for i, output := range effect.Behavior.Outputs {
			scaledOutputs[i] = output
			scaledOutputs[i].Amount = output.Amount * increase
		}

		log.Debug("Passive effect triggered",
			zap.String("card_name", effect.CardName),
			zap.String("trigger_type", trigger.Condition.Type),
			zap.String("resource_type", event.ResourceType),
			zap.Int("increase", increase))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID).
			WithCardRegistry(cr).
			WithSourceType(game.SourceTypePassiveEffect).
			WithOnCardsAddedToHand(MakeCardDrawCallback(p, g, cr))
		if err := applier.ApplyOutputs(context.Background(), scaledOutputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("Subscribed passive effect to ProductionChangedEvent",
		zap.String("card_name", effect.CardName))

	return subID
}

// createPassiveBehaviorChoice creates a pending behavior choice selection from a passive effect
// Used for effects like Viral Enhancers and Olympus Conference that require player to choose between options
func createPassiveBehaviorChoice(p *player.Player, effect player.CardEffect, log *zap.Logger) {
	p.Selection().SetPendingBehaviorChoiceSelection(&player.PendingBehaviorChoiceSelection{
		Choices:      effect.Behavior.Choices,
		Source:       effect.CardName,
		SourceCardID: effect.CardID,
	})

	log.Debug("Created pending behavior choice selection from passive effect",
		zap.String("card_name", effect.CardName),
		zap.Int("num_choices", len(effect.Behavior.Choices)))
}
