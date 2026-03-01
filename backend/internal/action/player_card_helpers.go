package action

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"

	"go.uber.org/zap"
)

// CreateAndCachePlayerCard creates a PlayerCard with event listeners and initial state,
// then caches it in the player's hand. This is called by actions when adding a card to hand.
func CreateAndCachePlayerCard(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) *player.PlayerCard {
	// 1. Create PlayerCard data holder
	pc := player.NewPlayerCard(card)

	// 2. Register event listeners for state recalculation
	registerPlayerCardEventListeners(pc, p, g, cardRegistry)

	// 3. Calculate initial state
	recalculatePlayerCard(pc, p, g, cardRegistry)

	// 4. Cache in Hand
	p.Hand().AddPlayerCard(card.ID, pc)

	return pc
}

// registerPlayerCardEventListeners registers all event listeners on a PlayerCard.
// Stores unsubscribe functions in PlayerCard for cleanup when card is removed.
func registerPlayerCardEventListeners(
	pc *player.PlayerCard,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	eventBus := g.EventBus()

	// When player resources change, recalculate affordability
	subID1 := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID1) })

	// When temperature changes, recalculate requirements
	subID2 := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID2) })

	// When oxygen changes, recalculate requirements
	subID3 := events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID3) })

	// When oceans change, recalculate requirements
	subID4 := events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID4) })

	// When player effects change (requirement modifiers), recalculate cost
	subID5 := events.Subscribe(eventBus, func(event events.PlayerEffectsChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID5) })

	// When game phase changes, recalculate state (affects phase validation)
	subID6 := events.Subscribe(eventBus, func(event events.GamePhaseChangedEvent) {
		if event.GameID == g.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID6) })

	// When general game state changes, recalculate availability
	subID7 := events.Subscribe(eventBus, func(event events.GameStateChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID7) })

	subID8 := events.Subscribe(eventBus, func(event events.CardPlayedEvent) {
		if event.GameID == g.ID() && event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID8) })

	subID9 := events.Subscribe(eventBus, func(event events.ProductionChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID9) })

	subID10 := events.Subscribe(eventBus, func(event events.TilePlacedEvent) {
		if event.GameID == g.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID10) })
}

// recalculatePlayerCard recalculates and updates PlayerCard state.
// Called on initial creation and by event listeners.
func recalculatePlayerCard(
	pc *player.PlayerCard,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	card, ok := pc.Card().(*gamecards.Card)
	if !ok {
		// Should never happen if architecture is followed correctly
		return
	}
	state := CalculatePlayerCardState(card, p, g, cardRegistry)
	pc.UpdateState(state)
}

// MakeCardDrawCallback returns a callback for the BehaviorApplier that creates PlayerCard
// caches when cards are drawn directly to hand via card-draw outputs.
func MakeCardDrawCallback(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) func(cardIDs []string) {
	return func(cardIDs []string) {
		for _, cardID := range cardIDs {
			card, err := cardRegistry.GetByID(cardID)
			if err != nil {
				continue
			}
			CreateAndCachePlayerCard(card, p, g, cardRegistry)
		}
	}
}

// AddCardsToPlayerHand adds multiple cards to a player's hand and creates PlayerCard instances.
// This is a convenience function that consolidates the common pattern of:
// 1. Adding card ID to hand (triggers events)
// 2. Fetching card from registry
// 3. Creating and caching PlayerCard with state and event listeners
//
// Used by actions that give cards to players (card draws, purchases, admin commands, etc.)
func AddCardsToPlayerHand(
	cardIDs []string,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) {
	for _, cardID := range cardIDs {
		// First add card ID to hand (triggers CardHandUpdatedEvent)
		p.Hand().AddCard(cardID)

		// Then create PlayerCard with state and event listeners, cache in hand
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			logger.Warn("Failed to get card from registry, skipping PlayerCard creation",
				zap.String("card_id", cardID),
				zap.Error(err))
			continue
		}
		CreateAndCachePlayerCard(card, p, g, cardRegistry)
	}
}
