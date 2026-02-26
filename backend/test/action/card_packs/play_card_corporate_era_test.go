package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// =============================================================================
// Card 006: Inventors' Guild (active, cost 9, tags: [science])
// "Action: look at the top card and either buy it or discard it"
// Behavior: manual trigger, outputs: card-buy 1 + card-peek 1 to self-player
// =============================================================================

func TestInventorsGuild_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-inventors-guild-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Inventors' Guild",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardBuy, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceCardPeek, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(cardID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Inventors' Guild should be in played cards")
}

func TestInventorsGuild_ActionCanBeUsed(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-inventors-guild-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Inventors' Guild",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardBuy, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceCardPeek, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 4)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(cardID)

	// Play the card first (registers the manual action)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild should play successfully")

	// Use the card action (behavior index 0 since only one behavior, manual)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild action should succeed")
}

// =============================================================================
// Card 014: Development Center (active, cost 11, tags: [building, science])
// "Action: Spend 1 energy to draw a card."
// Behavior: manual trigger, inputs: energy 1, outputs: card-draw 1
// =============================================================================

func TestDevelopmentCenter_PlaysAndActionSpendEnergyDrawCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-development-center-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Development Center",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 4)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourceEnergy: 3,
	})
	p.Hand().AddCard(cardID)

	energyBefore := p.Resources().Get().Energy

	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Development Center should be in played cards")

	// Use the card action: spend 1 energy to draw a card
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center action should succeed")

	energyAfter := p.Resources().Get().Energy
	testutil.AssertEqual(t, energyBefore-1, energyAfter,
		"Should have 1 less energy after using Development Center action")
}

func TestDevelopmentCenter_ActionFailsWithoutEnergy(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-development-center-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Development Center",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 4)

	// Give credits to play the card but NO energy
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(cardID)

	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center should play successfully")

	// Try to use the action with 0 energy - should fail
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Development Center action should fail without energy")
}

// --- Space Station (025) ---
// active, cost 10, tags: [space]
// "Effect: When you play a space card, you pay 2 M€ less for it."
// Auto trigger, outputs: discount 2 to self-player with selector tags:[space].

func TestSpaceStation_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-space-station-test",
		Name: "Space Station",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceDiscount,
						Amount:       2,
						Target:       "self-player",
						Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}},
					},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-space-station-test")

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-space-station-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Space Station should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-space-station-test"),
		"Space Station should be in played cards")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-10, creditsAfter,
		"Should have paid 10 credits for Space Station")
}

// --- Virus (050) ---
// event, cost 1, tags: [microbe]
// "Remove up to 2 animals or 5 plants from any player."
// Auto trigger with choices:
//   choice 0 = animal removal (2) from any-card
//   choice 1 = plant removal (5) from any-player

func TestVirus_Choice1_RemovePlantsFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-virus-test",
		Name: "Virus",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceAnimal, Amount: 2, Target: "any-card"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourcePlant, Amount: 5, Target: "any-player"},
						},
					},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 8,
	})
	attacker.Hand().AddCard("card-virus-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-virus-test", payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with choice 1 (remove plants)")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 3, targetResources.Plants, "Target should have 3 plants (8 - 5)")
}

func TestVirus_Choice1_PartialPlantRemoval(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-virus-test",
		Name: "Virus",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceAnimal, Amount: 2, Target: "any-card"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourcePlant, Amount: 5, Target: "any-player"},
						},
					},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 2,
	})
	attacker.Hand().AddCard("card-virus-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-virus-test", payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with partial plant removal")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 0, targetResources.Plants, "Target should have 0 plants (had 2, Virus removes up to 5)")
}

func TestVirus_Choice0_RemoveAnimalsFromCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-virus-test",
		Name: "Virus",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceAnimal, Amount: -2, Target: "any-card"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourcePlant, Amount: 5, Target: "any-player"},
						},
					},
				},
			},
		},
	}

	animalHost := gamecards.Card{
		ID:              "card-animal-host-virus",
		Name:            "Animal Host",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}

	additionalCards := []gamecards.Card{card, animalHost}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	p.PlayedCards().AddCard("card-animal-host-virus", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host-virus", 5)

	p.Hand().AddCard("card-virus-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-animal-host-virus"
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-virus-test", payment, &choiceIndex, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with choice 0 (remove animals)")

	animalStorage := p.Resources().GetCardStorage("card-animal-host-virus")
	testutil.AssertEqual(t, 3, animalStorage, "Animal host should have 3 animals (5 - 2)")
}

// --- Electro Catapult (069) ---
// Active card, cost 17, tags: [building]. Requirements: oxygen max 8.
// Behavior 0 (auto): outputs energy-production -1 to self-player
// Behavior 1 (manual): choices: [spend 1 plant, spend 1 steel] -> outputs credit 7 to self-player

func electroCatapultCard() gamecards.Card {
	return gamecards.Card{
		ID:   "card-electro-catapult-test",
		Name: "Electro Catapult",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 17,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Max: testutil.IntPtr(8)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"},
				},
				Choices: []shared.Choice{
					{
						Inputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
						},
					},
					{
						Inputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
						},
					},
				},
			},
		},
	}
}

func TestElectroCatapult_PlayDecreasesEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := electroCatapultCard()
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-electro-catapult-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-electro-catapult-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult should play successfully at 0% oxygen")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1 after playing Electro Catapult")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-electro-catapult-test"),
		"Electro Catapult should be in played cards")
}

func TestElectroCatapult_ActionSpendPlantGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-electro-catapult"
	p.PlayedCards().AddCard(cardID, "Electro Catapult", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 5,
	})
	creditsBefore := p.Resources().Get().Credits

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"},
		},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Electro Catapult",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	choiceIndex := 0 // spend 1 plant
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult action (spend plant) should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 4, resources.Plants, "Should have 4 plants after spending 1")
	testutil.AssertEqual(t, creditsBefore+7, resources.Credits, "Should gain 7 credits")
}

func TestElectroCatapult_ActionSpendSteelGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-electro-catapult-steel"
	p.PlayedCards().AddCard(cardID, "Electro Catapult", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 3,
	})
	creditsBefore := p.Resources().Get().Credits

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"},
		},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Electro Catapult",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	choiceIndex := 1 // spend 1 steel
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult action (spend steel) should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Steel, "Should have 2 steel after spending 1")
	testutil.AssertEqual(t, creditsBefore+7, resources.Credits, "Should gain 7 credits")
}

// --- Earth Catapult (070) ---
// Active card, cost 23, tags: [earth].
// "When you play a card, you pay 2 M€ less."
// Auto trigger, outputs: discount 2 to self-player (no selectors = applies to all cards).

func TestEarthCatapult_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-earth-catapult-test",
		Name: "Earth Catapult",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 23,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-earth-catapult-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-earth-catapult-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Earth Catapult should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-earth-catapult-test"),
		"Earth Catapult should be in played cards")
}

// =============================================================================
// Card 071: Advanced Alloys (active, cost 9, tags: [science])
// "Each steel/titanium worth 1 M€ extra."
// Auto trigger, outputs: 2x value-modifier (1 for steel, 1 for titanium) to
// self-player with selectors.
// =============================================================================

func TestAdvancedAlloys_PlaysSuccessfully_ValueModifiersApplied(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-advanced-alloys-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Advanced Alloys",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceValueModifier,
						Amount:       1,
						Target:       "self-player",
						Selectors: []shared.Selector{
							{Resources: []string{"steel"}},
						},
					},
					{
						ResourceType: shared.ResourceValueModifier,
						Amount:       1,
						Target:       "self-player",
						Selectors: []shared.Selector{
							{Resources: []string{"titanium"}},
						},
					},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(cardID)

	steelModBefore := p.Resources().GetValueModifier(shared.ResourceSteel)
	titaniumModBefore := p.Resources().GetValueModifier(shared.ResourceTitanium)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Advanced Alloys should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Advanced Alloys should be in played cards")

	steelModAfter := p.Resources().GetValueModifier(shared.ResourceSteel)
	titaniumModAfter := p.Resources().GetValueModifier(shared.ResourceTitanium)

	testutil.AssertEqual(t, steelModBefore+1, steelModAfter,
		"Steel value modifier should increase by 1")
	testutil.AssertEqual(t, titaniumModBefore+1, titaniumModAfter,
		"Titanium value modifier should increase by 1")
}

// =============================================================================
// Card 064: Mining Area (automated, cost 4, tags: [building])
// "Place mining tile on steel/titanium bonus area adjacent to your tile.
//  Increase production of that resource."
// Auto trigger with choices AND tile-placement output.
// Choice 0: steel-production +1, Choice 1: titanium-production +1.
// Plus outputs: tile-placement amount 1 with tileType "mining".
// =============================================================================

func TestMiningArea_Choice0_SteelProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-mining-area-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Mining Area",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
						},
					},
				},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceTilePlacement,
						Amount:       1,
						Target:       "none",
						TileType:     "mining",
						TileRestrictions: &shared.TileRestrictions{
							OnBonusType:     []string{"steel", "titanium"},
							AdjacentToOwned: true,
						},
					},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Place a player-owned tile adjacent to the steel bonus tile at (-3, 1, 2).
	// Neighbor of (-3, 1, 2) is (-2, 1, 1) which is a land tile on the standard board.
	adjacentToSteelBonus := shared.HexPosition{Q: -2, R: 1, S: 1}
	err := testGame.Board().UpdateTileOccupancy(ctx, adjacentToSteelBonus,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "Should place player tile adjacent to steel bonus")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(cardID)

	prodBefore := p.Resources().Production()

	choiceIndex := 0
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mining Area with choice 0 (steel production) should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Mining Area should be in played cards")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+1, prodAfter.Steel,
		"Steel production should increase by 1 when choosing choice 0")
}

func TestMiningArea_Choice1_TitaniumProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-mining-area-ti-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Mining Area",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
						},
					},
				},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceTilePlacement,
						Amount:       1,
						Target:       "none",
						TileType:     "mining",
						TileRestrictions: &shared.TileRestrictions{
							OnBonusType:     []string{"steel", "titanium"},
							AdjacentToOwned: true,
						},
					},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Place a player-owned tile adjacent to the titanium bonus tile at (2, 1, -3).
	// Neighbor of (2, 1, -3) is (1, 1, -2) which is a land tile on the standard board.
	adjacentToTitaniumBonus := shared.HexPosition{Q: 1, R: 1, S: -2}
	err := testGame.Board().UpdateTileOccupancy(ctx, adjacentToTitaniumBonus,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "Should place player tile adjacent to titanium bonus")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(cardID)

	prodBefore := p.Resources().Production()

	choiceIndex := 1
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mining Area with choice 1 (titanium production) should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Mining Area should be in played cards")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1 when choosing choice 1")
}

// --- Mars University (073) ---
// "When you play a science tag, including this, you may discard a card from hand to draw a card."
// Passive triggered effect: auto trigger with condition type:"tag-played" for science tags.
// Optional card-discard input, card-draw output. Just test that the card plays successfully.

func TestMarsUniversity_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-mars-university-test",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
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
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDiscard, Amount: 1, Target: "self-player", Optional: true},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mars-university-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mars-university-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mars University should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-mars-university-test"),
		"Mars University should be in played cards")
}

// --- Viral Enhancers (074) ---
// "When you play a plant, microbe, or an animal tag, including this, gain 1 plant or add 1 resource to that card."
// Passive triggered effect with choices. Just test that the card plays successfully.

func TestViralEnhancers_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-viral-enhancers-test",
		Name: "Viral Enhancers",
		Type: gamecards.CardTypeActive,
		Cost: 9,
		Tags: []shared.CardTag{shared.TagMicrobe, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{
					{
						Type: shared.TriggerTypeAuto,
						Condition: &shared.ResourceTriggerCondition{
							Type: "tag-played",
							Selectors: []shared.Selector{
								{Tags: []shared.CardTag{shared.TagPlant}},
								{Tags: []shared.CardTag{shared.TagMicrobe}},
								{Tags: []shared.CardTag{shared.TagAnimal}},
							},
						},
					},
				},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "any-card"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "any-card"},
						},
					},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-viral-enhancers-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-viral-enhancers-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Viral Enhancers should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-viral-enhancers-test"),
		"Viral Enhancers should be in played cards")
}

// --- Robotic Workforce (086) ---
// "Duplicate only the production box of one of your building cards."
// Has NO behaviors in JSON (null). Just test it plays successfully.

func TestRoboticWorkforce_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-robotic-workforce-test",
		Name: "Robotic Workforce",
		Type: gamecards.CardTypeAutomated,
		Cost: 9,
		Tags: []shared.CardTag{shared.TagScience},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-robotic-workforce-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-robotic-workforce-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Robotic Workforce should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-robotic-workforce-test"),
		"Robotic Workforce should be in played cards")
}

// --- Earth Office (105) ---
// Active card, cost 1, tags: [earth].
// "When you play an Earth tag, you pay 3 M€ less."
// Auto trigger, outputs: discount 3 to self-player with selector tags:[earth].

func TestEarthOffice_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-earth-office-test",
		Name: "Earth Office",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceDiscount,
						Amount:       3,
						Target:       "self-player",
						Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagEarth}}},
					},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-earth-office-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-earth-office-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Earth Office should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-earth-office-test"),
		"Earth Office should be in played cards")
}

// --- Business Contacts (111) ---
// Event card, cost 7, tags: [earth].
// "Look at top 4 cards, take 2, discard 2."
// Auto trigger, outputs: card-take 2 + card-peek 4 to self-player.

func TestBusinessContacts_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-business-contacts-test",
		Name: "Business Contacts",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 7,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardTake, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCardPeek, Amount: 4, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-business-contacts-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-business-contacts-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Business Contacts should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-business-contacts-test"),
		"Business Contacts should be in played cards")
}

// --- Sabotage (121) ---
// Event card, cost 1, tags: none.
// "Remove up to 3 titanium from any player, or 4 steel, or 7 M€."
// Auto trigger with 3 choices targeting any-player.

func TestSabotage_RemoveTitaniumFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-sabotage-test",
		Name: "Sabotage",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceTitanium, Amount: 3, Target: "any-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceSteel, Amount: 4, Target: "any-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceCredit, Amount: 7, Target: "any-player"},
						},
					},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-sabotage-test")

	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 5,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	choiceIndex := 0
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-sabotage-test", payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Sabotage should play successfully with choice 0 (remove titanium)")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Titanium, "Target should have 2 titanium after 3 removed (5 - 3)")
}

func TestSabotage_RemoveSteelFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-sabotage-steel-test",
		Name: "Sabotage",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 1,
		Tags: []shared.CardTag{},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Choices: []shared.Choice{
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceTitanium, Amount: 3, Target: "any-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceSteel, Amount: 4, Target: "any-player"},
						},
					},
					{
						Outputs: []shared.ResourceCondition{
							{ResourceType: shared.ResourceCredit, Amount: 7, Target: "any-player"},
						},
					},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-sabotage-steel-test")

	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 6,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	choiceIndex := 1
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-sabotage-steel-test", payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Sabotage should play successfully with choice 1 (remove steel)")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Steel, "Target should have 2 steel after 4 removed (6 - 4)")
}

// --- CEO's Favorite Project (149) ---
// "Add 1 resource to a card with at least 1 resource on it."
// Event, cost 1, no tags. Auto trigger, outputs: card-resource 1 to any-card.

func TestCEOsFavoriteProject_AddsMicrobeToTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite-project",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardResource, Amount: 1, Target: "any-card"},
				},
			},
		},
	}

	targetCard := gamecards.Card{
		ID:              "card-target-microbes",
		Name:            "Target Microbe Card",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagMicrobe},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, targetCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Set up a played card with microbe storage containing 2 microbes
	p.PlayedCards().AddCard("card-target-microbes", "Target Microbe Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("card-target-microbes", 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite-project")

	storageBefore := p.Resources().GetCardStorage("card-target-microbes")
	testutil.AssertEqual(t, 2, storageBefore, "Target card should start with 2 microbes")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-target-microbes"
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite-project", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "CEO's Favorite Project should play successfully")

	storageAfter := p.Resources().GetCardStorage("card-target-microbes")
	testutil.AssertEqual(t, 3, storageAfter, "Target card should have 3 microbes (2 existing + 1 added)")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-ceos-favorite-project"),
		"CEO's Favorite Project should be in played cards")
	testutil.AssertEqual(t, false, p.Hand().HasCard("card-ceos-favorite-project"),
		"CEO's Favorite Project should be removed from hand")
}

func TestCEOsFavoriteProject_FailsWithoutTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ceosFavorite := gamecards.Card{
		ID:   "card-ceos-favorite-project",
		Name: "CEO's Favorite Project",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardResource, Amount: 1, Target: "any-card"},
				},
			},
		},
	}

	targetCard := gamecards.Card{
		ID:              "card-target-animals",
		Name:            "Target Animal Card",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, targetCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("card-target-animals", "Target Animal Card", "active", []string{"animal"})
	p.Resources().AddToStorage("card-target-animals", 1)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ceos-favorite-project")

	// Play without specifying a target card - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ceos-favorite-project", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail without target card for card-resource output")
}

// --- Protected Habitats (173) ---
// "Opponents may not remove your plants/animals/microbes."
// Active, cost 5, no tags. Auto trigger, outputs: defense 1 to self-card with selectors for plant, microbe, animal.

func TestProtectedHabitats_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	protectedHabitats := gamecards.Card{
		ID:   "card-protected-habitats",
		Name: "Protected Habitats",
		Type: gamecards.CardTypeActive,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceDefense,
						Amount:       1,
						Target:       "self-card",
						Selectors: []shared.Selector{
							{Resources: []string{"plant"}},
							{Resources: []string{"microbe"}},
							{Resources: []string{"animal"}},
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{protectedHabitats})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-protected-habitats")

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-protected-habitats", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Protected Habitats should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-protected-habitats"),
		"Protected Habitats should be in played cards")
	testutil.AssertEqual(t, false, p.Hand().HasCard("card-protected-habitats"),
		"Protected Habitats should be removed from hand")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-5, creditsAfter,
		"Should have paid 5 credits for Protected Habitats")
}

// --- Corporate Stronghold (182) ---
// "Decrease your energy production 1 step and increase your M€ production 3 steps. Place a city tile."

func TestCorporateStronghold_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-corporate-stronghold-test",
		Name: "Corporate Stronghold",
		Type: gamecards.CardTypeAutomated,
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-corporate-stronghold-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-corporate-stronghold-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Corporate Stronghold should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

func TestCorporateStronghold_FailsWithoutEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-corporate-stronghold-test",
		Name: "Corporate Stronghold",
		Type: gamecards.CardTypeAutomated,
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-corporate-stronghold-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-corporate-stronghold-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Corporate Stronghold should fail without energy production")
}

// --- Olympus Conference (185) ---
// "When you play a science tag, including this, either add a science resource to this card,
// or remove a science resource from this card to draw a card."
// Passive triggered effect with choices and resource storage. Just test that the card plays successfully.

func TestOlympusConference_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-olympus-conference-test",
		Name: "Olympus Conference",
		Type: gamecards.CardTypeActive,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagEarth, shared.TagScience},
		ResourceStorage: &gamecards.ResourceStorage{
			Type: shared.ResourceScience,
		},
		Behaviors: []shared.CardBehavior{
			{
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
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-olympus-conference-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-olympus-conference-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Olympus Conference should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-olympus-conference-test"),
		"Olympus Conference should be in played cards")
}

// --- Invention Contest (192) ---
// "Look at top 3 cards, take 1, discard 2."
// Auto trigger, outputs: card-take 1 + card-peek 3.

func TestInventionContest_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-invention-contest-test",
		Name: "Invention Contest",
		Type: gamecards.CardTypeEvent,
		Cost: 2,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardTake, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceCardPeek, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-invention-contest-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-invention-contest-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Invention Contest should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-invention-contest-test"),
		"Invention Contest should be in played cards")
}

// --- Power Infrastructure (194) ---
// "Action: Spend any number of energy to gain that amount of M€."
// Manual trigger, variableAmount inputs (energy) and outputs (credit).

func TestPowerInfrastructure_PlayAndUseAction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-power-infrastructure-test",
		Name: "Power Infrastructure",
		Type: gamecards.CardTypeActive,
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourceEnergy: 10,
	})
	p.Hand().AddCard("card-power-infrastructure-test")

	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-power-infrastructure-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-power-infrastructure-test"),
		"Power Infrastructure should be in played cards")

	creditsBefore := p.Resources().Get().Credits
	energyBefore := p.Resources().Get().Energy

	// Give player another action since playing the card consumed one
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Use the action: spend 3 energy to gain 3 credits
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 3
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), "card-power-infrastructure-test", 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, energyBefore-3, resources.Energy, "Energy should decrease by 3")
	testutil.AssertEqual(t, creditsBefore+3, resources.Credits, "Credits should increase by 3")
}

func TestPowerInfrastructure_UseActionSpendAllEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-power-infra-ce9"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 7,
		shared.ResourceCredit: 5,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Spend all 7 energy
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 7
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure should succeed spending all energy")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Energy, "Should have 0 energy after spending all 7")
	testutil.AssertEqual(t, 12, resources.Credits, "Should have 12 credits (5 + 7)")
}

// --- Indentured Workers (195) ---
// "The next card you play this generation costs 8 M€ less."
// Auto trigger, outputs: discount 8 with temporary:"next-card".

func TestIndenturedWorkers_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-indentured-workers-test",
		Name: "Indentured Workers",
		Type: gamecards.CardTypeEvent,
		Cost: 0,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 8, Target: "self-player", Temporary: shared.TemporaryNextCard},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-indentured-workers-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-indentured-workers-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Indentured Workers should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-indentured-workers-test"),
		"Indentured Workers should be in played cards")
}

// =============================================================================
// Card B11: Saturn Systems (corporation, cost 0, tags: [jovian])
// "Effect: Each time any Jovian tag is put into play, including this,
//  increase your M€ production 1 step. You start with 1 titanium production
//  and 42 M€."
//
// Behavior 0: auto-corporation-start trigger, outputs credit 42 + titanium-production 1
// Behavior 1: auto trigger with condition tag-played for jovian (target any-player),
//             outputs credit-production 1
//
// When played as a regular card via PlayCardAction:
//   - Behavior 0 (auto-corporation-start) is NOT processed by PlayCardAction
//   - Behavior 1 (auto with condition) is registered as a passive effect
// =============================================================================

func TestSaturnSystems_CorporationPlays(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	cardID := "card-saturn-systems-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Saturn Systems",
		Type: gamecards.CardTypeCorporation,
		Pack: "corporate-era",
		Cost: 0,
		Tags: []shared.CardTag{shared.TagJovian},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: "auto-corporation-start"}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 42, Target: "self-player"},
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{
					{
						Type: shared.TriggerTypeAuto,
						Condition: &shared.ResourceTriggerCondition{
							Type: "tag-played",
							Selectors: []shared.Selector{
								{Tags: []shared.CardTag{shared.TagJovian}},
							},
							Target: &anyPlayerTarget,
						},
					},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(cardID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Saturn Systems should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Saturn Systems should be in played cards")

	// The auto-corporation-start behavior is NOT processed by PlayCardAction,
	// so credits and titanium production should be unchanged from the play action itself.
	// However, the conditional trigger (behavior 1) should be registered as a passive effect.
	effects := p.Effects().List()
	foundPassiveEffect := false
	for _, effect := range effects {
		if effect.CardID == cardID && effect.CardName == "Saturn Systems" {
			foundPassiveEffect = true
			testutil.AssertEqual(t, 1, effect.BehaviorIndex,
				"Passive effect should reference behavior index 1")
			break
		}
	}
	testutil.AssertTrue(t, foundPassiveEffect,
		"Saturn Systems should have registered its Jovian tag-played passive effect")
}

// =============================================================================
// Card B12: Teractor (corporation, cost 0, tags: [earth])
// "Effect: When playing an Earth card, you pay 3 M€ less for it.
//  You start with 60 M€."
//
// Behavior 0: auto-corporation-start trigger, outputs credit 60
// Behavior 1: auto trigger (no condition), outputs discount 3
//             with selector tags:[earth]
//
// When played as a regular card via PlayCardAction:
//   - Behavior 0 (auto-corporation-start) is NOT processed by PlayCardAction
//   - Behavior 1 (auto, no condition, persistent discount) is applied immediately
//     and registered as a persistent effect
// =============================================================================

func TestTeractor_CorporationPlays(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-teractor-test"
	card := gamecards.Card{
		ID:   cardID,
		Name: "Teractor",
		Type: gamecards.CardTypeCorporation,
		Pack: "corporate-era",
		Cost: 0,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: "auto-corporation-start"}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 60, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceDiscount,
						Amount:       3,
						Target:       "self-player",
						Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagEarth}}},
					},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(cardID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Teractor should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains(cardID),
		"Teractor should be in played cards")

	// The auto behavior (behavior 1) with persistent discount output should be
	// registered as an effect. Verify the discount effect exists.
	effects := p.Effects().List()
	foundDiscountEffect := false
	for _, effect := range effects {
		if effect.CardID == cardID && effect.CardName == "Teractor" {
			foundDiscountEffect = true
			testutil.AssertEqual(t, 1, effect.BehaviorIndex,
				"Discount effect should reference behavior index 1")
			// Verify the behavior contains the expected discount output
			testutil.AssertEqual(t, 1, len(effect.Behavior.Outputs),
				"Discount behavior should have 1 output")
			testutil.AssertEqual(t, shared.ResourceDiscount, effect.Behavior.Outputs[0].ResourceType,
				"Output should be a discount resource type")
			testutil.AssertEqual(t, 3, effect.Behavior.Outputs[0].Amount,
				"Discount should be 3 M€")
			break
		}
	}
	testutil.AssertTrue(t, foundDiscountEffect,
		"Teractor should have registered its Earth card discount effect")

	// Verify the discount is actually calculated for Earth-tagged cards
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	earthCard := &gamecards.Card{
		ID:   "card-earth-tag-test",
		Name: "Test Earth Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagEarth},
	}
	discount := calculator.CalculateCardDiscounts(p, earthCard)
	testutil.AssertEqual(t, 3, discount,
		"Earth card should receive 3 M€ discount from Teractor")

	// Verify non-Earth cards do NOT get the discount
	nonEarthCard := &gamecards.Card{
		ID:   "card-non-earth-test",
		Name: "Test Non-Earth Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	noDiscount := calculator.CalculateCardDiscounts(p, nonEarthCard)
	testutil.AssertEqual(t, 0, noDiscount,
		"Non-Earth card should not receive any discount from Teractor")
}
