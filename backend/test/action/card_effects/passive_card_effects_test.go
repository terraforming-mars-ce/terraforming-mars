package card_effects_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	gamecards "terraforming-mars-backend/internal/game/cards"
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	other.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// Register Rover Construction passive effect
	effect := shared.CardEffect{
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	effect := shared.CardEffect{
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	effect := shared.CardEffect{
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// Set up Olympus Conference as played with resource storage
	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	// Register tag-played passive effect (choice 0: add science to self-card)
	effect := shared.CardEffect{
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	effect := shared.CardEffect{
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

// --- Olympus Conference (185) - Triggered Choice Tests ---

func olympusConferenceChoiceBehavior() shared.CardBehavior {
	return shared.CardBehavior{
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
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
}

func TestOlympusConference_TriggeredChoice_CreatesPendingSelection(t *testing.T) {
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	effect := shared.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior:      olympusConferenceChoiceBehavior(),
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	testutil.AssertTrue(t, owner.Selection().GetPendingBehaviorChoiceSelection() == nil,
		"Should have no pending selection before event")

	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	selection := owner.Selection().GetPendingBehaviorChoiceSelection()
	testutil.AssertTrue(t, selection != nil, "Should have a pending behavior choice selection")
	testutil.AssertEqual(t, 2, len(selection.Choices), "Should have 2 choices")
	testutil.AssertEqual(t, "Olympus Conference", selection.Source, "Source should be Olympus Conference")
	testutil.AssertEqual(t, "card-olympus-conference", selection.SourceCardID, "SourceCardID should match")
}

func TestOlympusConference_Choice0_AddScience(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	effect := shared.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior:      olympusConferenceChoiceBehavior(),
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	storageBefore := owner.Resources().GetCardStorage("card-olympus-conference")

	confirmBehaviorChoice := confirmAction.NewConfirmBehaviorChoiceAction(repo, cardRegistry, logger)
	err := confirmBehaviorChoice.Execute(ctx, testGame.ID(), owner.ID(), 0, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add science) should succeed")

	storageAfter := owner.Resources().GetCardStorage("card-olympus-conference")
	testutil.AssertEqual(t, storageBefore+1, storageAfter,
		"Should have 1 more science resource on card after choice 0")

	testutil.AssertTrue(t, owner.Selection().GetPendingBehaviorChoiceSelection() == nil,
		"Pending selection should be cleared after confirming")
}

func TestOlympusConference_Choice1_RemoveScienceToDrawCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 2)

	effect := shared.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior:      olympusConferenceChoiceBehavior(),
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	handBefore := owner.Hand().CardCount()

	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	confirmBehaviorChoice := confirmAction.NewConfirmBehaviorChoiceAction(repo, cardRegistry, logger)
	err := confirmBehaviorChoice.Execute(ctx, testGame.ID(), owner.ID(), 1, nil)
	testutil.AssertNoError(t, err, "Choice 1 (remove science, draw card) should succeed")

	storageAfter := owner.Resources().GetCardStorage("card-olympus-conference")
	testutil.AssertEqual(t, 1, storageAfter,
		"Should have 1 science resource after spending 1 from 2")

	handAfter := owner.Hand().CardCount()
	testutil.AssertEqual(t, handBefore+1, handAfter,
		"Should have drawn 1 card")

	testutil.AssertTrue(t, owner.Selection().GetPendingBehaviorChoiceSelection() == nil,
		"Pending selection should be cleared after confirming")
}

func TestOlympusConference_Choice1_FailsWithoutScience(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-olympus-conference", "Olympus Conference", "active", []string{"building", "earth", "science"})
	owner.Resources().AddToStorage("card-olympus-conference", 0)

	effect := shared.CardEffect{
		CardID:        "card-olympus-conference",
		CardName:      "Olympus Conference",
		BehaviorIndex: 0,
		Behavior:      olympusConferenceChoiceBehavior(),
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	confirmBehaviorChoice := confirmAction.NewConfirmBehaviorChoiceAction(repo, cardRegistry, logger)
	err := confirmBehaviorChoice.Execute(ctx, testGame.ID(), owner.ID(), 1, nil)
	testutil.AssertError(t, err, "Choice 1 should fail with 0 science resources on card")
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-viral-enhancers", "Viral Enhancers", "active", []string{"microbe", "science"})

	// Register card-played passive effect (simplified to just gain plant)
	effect := shared.CardEffect{
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
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.PlayedCards().AddCard("card-viral-enhancers", "Viral Enhancers", "active", []string{"microbe", "science"})

	effect := shared.CardEffect{
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

// --- Arctic Algae (023) ---
// "Effect: When anyone places an ocean tile, gain 2 plants."

func TestArcticAlgae_GainPlantsOnAnyOceanPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	arcticAlgaeCard := gamecards.Card{
		ID:   "card-arctic-algae",
		Name: "Arctic Algae",
		Type: gamecards.CardTypeActive,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagPlant},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{arcticAlgaeCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	other := players[1]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	other.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	effect := shared.CardEffect{
		CardID:        "card-arctic-algae",
		CardName:      "Arctic Algae",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:   "ocean-placed",
						Target: &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	plantsBefore := owner.Resources().Get().Plants

	// Another player places an ocean tile
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: other.ID(),
		TileType: string(shared.ResourceOceanTile),
	})

	time.Sleep(20 * time.Millisecond)

	plantsAfter := owner.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter,
		"Arctic Algae owner should gain 2 plants when any player places an ocean")
}

func TestArcticAlgae_TriggersOnSelfOceanToo(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	arcticAlgaeCard := gamecards.Card{
		ID:   "card-arctic-algae",
		Name: "Arctic Algae",
		Type: gamecards.CardTypeActive,
		Cost: 12,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{arcticAlgaeCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	effect := shared.CardEffect{
		CardID:        "card-arctic-algae",
		CardName:      "Arctic Algae",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:   "ocean-placed",
						Target: &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	plantsBefore := owner.Resources().Get().Plants

	// Owner places an ocean tile
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		TileType: string(shared.ResourceOceanTile),
	})

	time.Sleep(20 * time.Millisecond)

	plantsAfter := owner.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter,
		"Arctic Algae should also trigger when owner places an ocean")
}

func TestArcticAlgae_DoesNotTriggerOnCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	arcticAlgaeCard := gamecards.Card{
		ID:   "card-arctic-algae",
		Name: "Arctic Algae",
		Type: gamecards.CardTypeActive,
		Cost: 12,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{arcticAlgaeCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	effect := shared.CardEffect{
		CardID:        "card-arctic-algae",
		CardName:      "Arctic Algae",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type:   "ocean-placed",
						Target: &anyPlayerTarget,
					},
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	plantsBefore := owner.Resources().Get().Plants

	// Place a city (not an ocean)
	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		TileType: string(shared.ResourceCityTile),
	})

	time.Sleep(20 * time.Millisecond)

	plantsAfter := owner.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore, plantsAfter,
		"Arctic Algae should NOT trigger on city placement")
}
