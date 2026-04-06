package behavior_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// ============================================================================
// Card resource tests
// ============================================================================

// --- CEO's Favorite Project (149) ---
// "Add 1 resource to a card with at least 1 resource on it"

func TestCardResource_CEOsFavoriteProject_AddsToAnimalCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceCardResource, 1, "any-card"),
				},
			},
		},
	}

	animalCard := gamecards.Card{
		ID: "card-animal-host", Name: "Animal Host", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, animalCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	// Set up played card with animal storage (with 1 existing resource)
	p.PlayedCards().AddCard("card-animal-host", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host", 1) // Has 1 resource already

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite")

	// Play CEO's Favorite Project targeting the animal card
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-animal-host"
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "CEO's Favorite Project should play successfully")

	// Verify 1 resource was added (1 existing + 1 new = 2)
	storage := p.Resources().GetCardStorage("card-animal-host")
	testutil.AssertEqual(t, 2, storage, "Should have 2 animals on card (1 existing + 1 added)")
}

func TestCardResource_CEOsFavoriteProject_AddsToMicrobeCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceCardResource, 1, "any-card"),
				},
			},
		},
	}

	microbeCard := gamecards.Card{
		ID: "card-microbe-host", Name: "Microbe Host", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagMicrobe},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, microbeCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	// Set up played card with microbe storage
	p.PlayedCards().AddCard("card-microbe-host", "Microbe Host", "active", []string{"microbe"})
	p.Resources().AddToStorage("card-microbe-host", 3) // Has 3 microbes

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite")

	// Play CEO's Favorite Project targeting the microbe card
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-microbe-host"
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "CEO's Favorite Project should play targeting microbe card")

	// Verify 1 resource was added (3 existing + 1 new = 4)
	storage := p.Resources().GetCardStorage("card-microbe-host")
	testutil.AssertEqual(t, 4, storage, "Should have 4 microbes on card (3 existing + 1 added)")
}

// --- Corroder Suits (219) ---
// "Increase your M$ production 2 steps. Add 1 resource to **any venus card**."

func TestCardResource_CorroderSuits_AddsToVenusCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	corroderSuits := gamecards.Card{
		ID:   "card-corroder-suits",
		Name: "Corroder Suits",
		Type: gamecards.CardTypeAutomated,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagVenus},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceCreditProduction, 2, "self-player"),
					&shared.CardStorageCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardResource, Amount: 1, Target: "any-card"},
						Selectors:     []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}},
					},
				},
			},
		},
	}

	venusFloaterCard := gamecards.Card{
		ID: "card-venus-floater", Name: "Venus Floater Card", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagVenus},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceFloater, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{corroderSuits, venusFloaterCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	// Set up played venus card with floater storage
	p.PlayedCards().AddCard("card-venus-floater", "Venus Floater Card", "active", []string{"venus"})
	p.Resources().AddToStorage("card-venus-floater", 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-corroder-suits")

	productionBefore := p.Resources().Production()

	// Play Corroder Suits targeting the venus floater card
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	targetCardID := "card-venus-floater"
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-corroder-suits", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Corroder Suits should play successfully")

	// Verify production increased
	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+2, productionAfter.Credits, "Should gain 2 credit production")

	// Verify 1 resource was added to the venus card
	storage := p.Resources().GetCardStorage("card-venus-floater")
	testutil.AssertEqual(t, 1, storage, "Should have 1 floater on venus card")
}

// --- Maxwell Base (238) - Card Action ---
// "Action: Add 1 resource to **another venus card**."

func TestCardResource_MaxwellBase_ActionAddsToVenusCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	maxwellBase := gamecards.Card{
		ID:   "card-maxwell-base",
		Name: "Maxwell Base",
		Type: gamecards.CardTypeActive,
		Cost: 18,
		Tags: []shared.CardTag{shared.TagCity, shared.TagVenus},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceEnergyProduction, -1, "self-player"),
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.BehaviorCondition{
					&shared.CardStorageCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardResource, Amount: 1, Target: "any-card"},
						Selectors:     []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}},
					},
				},
			},
		},
	}

	venusMicrobeCard := gamecards.Card{
		ID: "card-venus-microbe", Name: "Venus Microbe Card", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagVenus},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{maxwellBase, venusMicrobeCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	// Set up the venus microbe card
	p.PlayedCards().AddCard("card-venus-microbe", "Venus Microbe Card", "active", []string{"venus"})
	p.Resources().AddToStorage("card-venus-microbe", 2) // 2 existing microbes

	// Register Maxwell Base as a played card with its manual action
	p.PlayedCards().AddCard("card-maxwell-base", "Maxwell Base", "active", []string{"city", "venus"})
	p.Actions().AddAction(shared.CardAction{
		CardID:        maxwellBase.ID,
		CardName:      maxwellBase.Name,
		BehaviorIndex: 1,
		Behavior:      maxwellBase.Behaviors[1],
	})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	// Use Maxwell Base action targeting the venus microbe card
	useCardAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	targetCardID := "card-venus-microbe"
	err = useCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-maxwell-base", 1, nil, []string{targetCardID}, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Maxwell Base action should execute successfully")

	// Verify 1 resource was added (2 existing + 1 new = 3)
	storage := p.Resources().GetCardStorage("card-venus-microbe")
	testutil.AssertEqual(t, 3, storage, "Should have 3 microbes on venus card (2 existing + 1 added)")
}

// --- card-resource fails without target ---

func TestCardResource_FailsWithoutTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceCardResource, 1, "any-card"),
				},
			},
		},
	}

	animalCard := gamecards.Card{
		ID: "card-animal-host", Name: "Animal Host", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, animalCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	p.PlayedCards().AddCard("card-animal-host", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host", 1)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite")

	// Play without specifying a target card — should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail without target card for card-resource output")
}

// --- card-resource fails when target card has no storage ---

func TestCardResource_FailsWhenTargetHasNoStorage(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceCardResource, 1, "any-card"),
				},
			},
		},
	}

	noStorageCard := gamecards.Card{
		ID: "card-no-storage", Name: "No Storage Card", Type: gamecards.CardTypeAutomated, Cost: 0,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, noStorageCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	p.PlayedCards().AddCard("card-no-storage", "No Storage Card", "automated", []string{})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite")

	// Play targeting a card with no storage — should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-no-storage"
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertError(t, err, "Should fail when target card has no resource storage")
}

// --- Hydrogen To Venus (231) ---
// "Raise Venus 1 step. Add 1 floater to a venus card for each Jovian tag you have."

func TestCardResource_AnyCardTarget_SkipsWhenNoTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	jovianTag := shared.TagJovian
	selfPlayer := "self-player"

	hydrogenToVenus := gamecards.Card{
		ID:   "card-hydrogen-to-venus",
		Name: "Hydrogen To Venus",
		Type: gamecards.CardTypeEvent,
		Cost: 11,
		Tags: []shared.CardTag{"space"},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceCreditProduction, 1, "self-player"),
					&shared.CardStorageCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceFloater, Amount: 1, Target: "any-card"},
						Selectors:     []shared.Selector{{Tags: []shared.CardTag{"venus"}}},
						Per: &shared.PerCondition{
							ResourceType: "tag",
							Amount:       1,
							Target:       &selfPlayer,
							Tag:          &jovianTag,
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{hydrogenToVenus})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")

	// Give player jovian tags (via played cards) but NO venus card with floater storage
	p.PlayedCards().AddCard("card-jovian-1", "Jovian Card", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("card-jovian-2", "Jovian Card 2", "automated", []string{"jovian"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-hydrogen-to-venus")

	initialProduction := p.Resources().Production().Credits

	// Play without providing a target card — should succeed, skipping floater placement
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-hydrogen-to-venus", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Card should play successfully even without a valid floater target")

	// Verify the other output (production) was still applied
	newProduction := p.Resources().Production().Credits
	testutil.AssertEqual(t, initialProduction+1, newProduction, "Credit production should have increased by 1")

	// Verify card was removed from hand
	testutil.AssertEqual(t, false, p.Hand().HasCard("card-hydrogen-to-venus"), "Card should be removed from hand")
}

// ============================================================================
// Card steal tests
// ============================================================================

func TestPredatorsStealAnimalFromOtherPlayer(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-birds"
	other.PlayedCards().AddCard(targetCardID, "Birds", "active", []string{"animal"})
	other.Resources().AddToStorage(targetCardID, 3)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceAnimal, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal action should succeed")

	testutil.AssertEqual(t, 2, other.Resources().GetCardStorage(targetCardID), "Target card should have 2 animals after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(predatorsID), "Predators card should have 1 animal after steal")
}

func TestPredatorsStealFromOwnCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-fish"
	p.PlayedCards().AddCard(targetCardID, "Fish", "active", []string{"animal"})
	p.Resources().AddToStorage(targetCardID, 2)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceAnimal, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal from own card should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(targetCardID), "Target card should have 1 animal after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(predatorsID), "Predators card should have 1 animal after steal")
}

func TestPredatorsRejectsWithNoSourceCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceAnimal, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Predators without source card should be rejected")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(predatorsID), "Predators card should have 0 animals when no source card specified")
}

func TestPredatorsStealFromCardWithZeroAnimals(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-birds"
	other.PlayedCards().AddCard(targetCardID, "Birds", "active", []string{"animal"})
	other.Resources().AddToStorage(targetCardID, 0)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceAnimal, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal from empty card should succeed")

	testutil.AssertEqual(t, 0, other.Resources().GetCardStorage(targetCardID), "Target card should still have 0 animals")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(predatorsID), "Predators card should have 0 animals when source had none")
}

func TestAntsStealMicrobeFromOtherPlayer(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	antsID := "test-ants"
	p.PlayedCards().AddCard(antsID, "Ants", "active", []string{"microbe"})
	p.Resources().AddToStorage(antsID, 0)

	targetCardID := "test-decomposers"
	other.PlayedCards().AddCard(targetCardID, "Decomposers", "active", []string{"microbe"})
	other.Resources().AddToStorage(targetCardID, 5)

	antsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceMicrobe, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        antsID,
			CardName:      "Ants",
			BehaviorIndex: 0,
			Behavior:      antsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, antsID, 0, nil, []string{antsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ants steal action should succeed")

	testutil.AssertEqual(t, 4, other.Resources().GetCardStorage(targetCardID), "Target card should have 4 microbes after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(antsID), "Ants card should have 1 microbe after steal")
}

func TestAntsRejectsWithNoSourceCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	antsID := "test-ants"
	p.PlayedCards().AddCard(antsID, "Ants", "active", []string{"microbe"})
	p.Resources().AddToStorage(antsID, 0)

	antsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			shared.NewCardStorageCondition(shared.ResourceMicrobe, 1, "steal-from-any-card"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        antsID,
			CardName:      "Ants",
			BehaviorIndex: 0,
			Behavior:      antsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, antsID, 0, nil, []string{antsID}, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Ants without source card should be rejected")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(antsID), "Ants card should have 0 microbes when no source card specified")
}

func TestCardStorage_VariableAmountInput(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-microbe-card"
	p.PlayedCards().AddCard(cardID, "Test Microbe Card", "active", []string{"microbe"})
	p.Resources().AddToStorage(cardID, 5)

	input := &shared.CardStorageCondition{
		ConditionBase:  shared.ConditionBase{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
		VariableAmount: true,
	}

	applier := gamecards.NewBehaviorApplier(p, testGame, "test", testutil.TestLogger()).
		WithSelectedAmount(2).
		WithSourceCardID(cardID)
	err := applier.ApplyInputs(context.Background(), []shared.BehaviorCondition{input})
	testutil.AssertNoError(t, err, "ApplyInputs should succeed")

	testutil.AssertEqual(t, 3, p.Resources().GetCardStorage(cardID), "Should have 3 microbes remaining after spending 2")
}

func TestCardStorage_StealFromCardWithZeroResources(t *testing.T) {
	testGame, _, _, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	// Player 2 has a card with animal storage but 0 animals
	otherCardID := "test-empty-animals"
	other.PlayedCards().AddCard(otherCardID, "Empty Animals", "active", []string{"animal"})
	other.Resources().AddToStorage(otherCardID, 0)

	// Player 1 has a card with animal storage and some animals
	selfCardID := "test-predators"
	p.PlayedCards().AddCard(selfCardID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(selfCardID, 3)

	output := shared.NewCardStorageCondition(shared.ResourceAnimal, 2, "steal-from-any-card")

	applyOutputsWithOptions(t, p, testGame, applyOptions{
		sourceCardID:      selfCardID,
		stealSourceCardID: otherCardID,
	}, output)

	// Nothing to steal from empty card
	testutil.AssertEqual(t, 0, other.Resources().GetCardStorage(otherCardID),
		"Other player's card should still have 0 animals")
	testutil.AssertEqual(t, 3, p.Resources().GetCardStorage(selfCardID),
		"Player's card should still have original 3 animals (nothing stolen)")
}
