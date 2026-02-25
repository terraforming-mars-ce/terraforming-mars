package card_effects_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Rover Construction (038) ---
// "Effect: When any city tile is placed, gain 2 M€."

func TestRoverConstruction_GainCreditsOnAnyCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	roverCard := gamecards.Card{
		ID:   "card-rover-construction",
		Name: "Rover Construction",
		Type: gamecards.CardTypeActive,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{roverCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	other := players[1]
	owner.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Register Rover Construction passive effect
	effect := player.CardEffect{
		CardID:        "card-rover-construction",
		CardName:      "Rover Construction",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:     "city-placed",
						Location: testutil.StrPtr("anywhere"),
						Target:   &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)

	// Subscribe the passive effect to events
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	creditsBefore := owner.Resources().Get().Credits

	// Simulate a city tile placement by another player
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: other.ID(),
		TileType: string(shared.ResourceCityTile),
	})

	time.Sleep(20 * time.Millisecond)

	creditsAfter := owner.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore+2, creditsAfter,
		"Rover Construction owner should gain 2 credits when any player places a city")
}

func TestRoverConstruction_TriggersOnSelfCityToo(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	roverCard := gamecards.Card{
		ID:   "card-rover-construction",
		Name: "Rover Construction",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{roverCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	effect := player.CardEffect{
		CardID:        "card-rover-construction",
		CardName:      "Rover Construction",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:     "city-placed",
						Location: testutil.StrPtr("anywhere"),
						Target:   &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	creditsBefore := owner.Resources().Get().Credits

	// Self places a city
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		TileType: string(shared.ResourceCityTile),
	})

	time.Sleep(20 * time.Millisecond)

	creditsAfter := owner.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore+2, creditsAfter,
		"Rover Construction should also trigger when owner places a city")
}

func TestRoverConstruction_DoesNotTriggerOnGreenery(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	roverCard := gamecards.Card{
		ID:   "card-rover-construction",
		Name: "Rover Construction",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{roverCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	effect := player.CardEffect{
		CardID:        "card-rover-construction",
		CardName:      "Rover Construction",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:     "city-placed",
						Location: testutil.StrPtr("anywhere"),
						Target:   &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	creditsBefore := owner.Resources().Get().Credits

	// Place a greenery (not a city)
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		TileType: string(shared.ResourceGreeneryTile),
	})

	time.Sleep(20 * time.Millisecond)

	creditsAfter := owner.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore, creditsAfter,
		"Rover Construction should NOT trigger on greenery placement")
}

// --- Olympus Conference (185) ---
// "Effect: When you play a science tag, including this, either add a science resource
//
//	to this card, or remove a science resource from this card to draw a card."

func TestOlympusConference_AddScienceOnScienceTagPlayed(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	olympusConference := gamecards.Card{
		ID:   "card-olympus-conference",
		Name: "Olympus Conference",
		Type: gamecards.CardTypeActive,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagEarth, shared.TagScience},
		ResourceStorage: &gamecards.ResourceStorage{
			Type:     shared.ResourceScience,
			Starting: 0,
		},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{olympusConference})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Set up Olympus Conference as played with resource storage
	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	// Register tag-played passive effect (choice 0: add science to self-card)
	effect := player.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type: "tag-played",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagScience}},
						},
					},
				},
			},
			// For testing: just the add-science output (choice 0)
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	storageBefore := owner.Resources().GetCardStorage("card-olympus-conference")

	// Simulate playing a science tag
	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	storageAfter := owner.Resources().GetCardStorage("card-olympus-conference")
	testutil.AssertEqual(t, storageBefore+1, storageAfter,
		"Olympus Conference should add 1 science resource when a science tag is played")
}

func TestOlympusConference_DoesNotTriggerOnNonScienceTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	olympusConference := gamecards.Card{
		ID:   "card-olympus-conference",
		Name: "Olympus Conference",
		Type: gamecards.CardTypeActive,
		Cost: 10,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{olympusConference})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	effect := player.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type: "tag-played",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagScience}},
						},
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	storageBefore := owner.Resources().GetCardStorage("card-olympus-conference")

	// Play a building tag (not science)
	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "building",
	})

	time.Sleep(20 * time.Millisecond)

	storageAfter := owner.Resources().GetCardStorage("card-olympus-conference")
	testutil.AssertEqual(t, storageBefore, storageAfter,
		"Olympus Conference should NOT trigger on non-science tag")
}

// --- Viral Enhancers (074) ---
// "Effect: When you play a plant, microbe, or an animal tag, including this,
//
//	gain 1 plant or add 1 resource to that card."

func TestViralEnhancers_GainPlantOnPlantTagPlayed(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	viralEnhancers := gamecards.Card{
		ID:   "card-viral-enhancers",
		Name: "Viral Enhancers",
		Type: gamecards.CardTypeActive,
		Cost: 9,
		Tags: []shared.CardTag{shared.TagMicrobe, shared.TagScience},
	}
	plantCard := gamecards.Card{
		ID:   "card-plant-test",
		Name: "Plant Test Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 0,
		Tags: []shared.CardTag{shared.TagPlant},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{viralEnhancers, plantCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-viral-enhancers", "Viral Enhancers", "active", []string{"microbe", "science"})

	// Register card-played passive effect (simplified to just gain plant)
	effect := player.CardEffect{
		CardID:        "card-viral-enhancers",
		CardName:      "Viral Enhancers",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:   "card-played",
						Target: testutil.StrPtr("self-player"),
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagPlant}},
							{Tags: []shared.CardTag{shared.TagMicrobe}},
							{Tags: []shared.CardTag{shared.TagAnimal}},
						},
					},
				},
			},
			// Simplified: just the "gain 1 plant" choice output for testing
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	plantsBefore := owner.Resources().Get().Plants

	// Simulate playing a plant-tagged card
	events.Publish(testGame.EventBus(), events.CardPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		CardID:   "card-plant-test",
		CardName: "Plant Test Card",
	})

	time.Sleep(20 * time.Millisecond)

	plantsAfter := owner.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+1, plantsAfter,
		"Viral Enhancers should give 1 plant when a plant-tagged card is played")
}

func TestViralEnhancers_DoesNotTriggerOnBuildingTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	viralEnhancers := gamecards.Card{
		ID:   "card-viral-enhancers",
		Name: "Viral Enhancers",
		Type: gamecards.CardTypeActive,
		Cost: 9,
		Tags: []shared.CardTag{shared.TagMicrobe, shared.TagScience},
	}
	buildingCard := gamecards.Card{
		ID:   "card-building-test",
		Name: "Building Test Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 0,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{viralEnhancers, buildingCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-viral-enhancers", "Viral Enhancers", "active", []string{"microbe", "science"})

	effect := player.CardEffect{
		CardID:        "card-viral-enhancers",
		CardName:      "Viral Enhancers",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:   "card-played",
						Target: testutil.StrPtr("self-player"),
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagPlant}},
							{Tags: []shared.CardTag{shared.TagMicrobe}},
							{Tags: []shared.CardTag{shared.TagAnimal}},
						},
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	plantsBefore := owner.Resources().Get().Plants

	// Play a building-tagged card (should NOT trigger)
	events.Publish(testGame.EventBus(), events.CardPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		CardID:   "card-building-test",
		CardName: "Building Test Card",
	})

	time.Sleep(20 * time.Millisecond)

	plantsAfter := owner.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore, plantsAfter,
		"Viral Enhancers should NOT trigger on building tag")
}
