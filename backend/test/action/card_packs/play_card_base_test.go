package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func nitriteReducingBacteriaBehavior() shared.CardBehavior {
	return shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 3, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTR, Amount: 1, Target: "none"},
				},
			},
		},
	}
}

// --- Colonizer Training Camp (001) ---
// No behaviors, just VP and a max oxygen requirement.
// Test that the card can be played when oxygen is low enough.

func TestColonizerTrainingCamp_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-colonizer-training-camp-test",
		Name: "Colonizer Training Camp",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagJovian},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Max: testutil.IntPtr(5)},
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
	p.Hand().AddCard("card-colonizer-training-camp-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-colonizer-training-camp-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Colonizer Training Camp should play successfully at 0% oxygen")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-colonizer-training-camp-test"),
		"Colonizer Training Camp should be in played cards")
}

// --- Deep Well Heating (003) ---
// "Increase your energy production 1 step. Increase temperature 1 step."

func TestDeepWellHeating_EnergyProductionAndTemperature(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-deep-well-heating-test",
		Name: "Deep Well Heating",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "none"},
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
	p.Hand().AddCard("card-deep-well-heating-test")

	prodBefore := p.Resources().Production()
	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-deep-well-heating-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Deep Well Heating should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+2, tempAfter,
		"Temperature should increase by 1 step (2 degrees)")
}

// --- Cloud Seeding (004) ---
// "Decrease your M€ production 1 step and any heat production 1 step. Increase your plant production 2 steps."

func TestCloudSeeding_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-cloud-seeding-test",
		Name: "Cloud Seeding",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	target := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	// Give player some credit production so it can decrease
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 2,
	})
	// Give target heat production
	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceHeatProduction: 3,
	})
	p.Hand().AddCard("card-cloud-seeding-test")

	prodBefore := p.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-cloud-seeding-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Cloud Seeding should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-1, prodAfter.Credits,
		"Credit production should decrease by 1")
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")

	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, targetProdBefore.Heat-1, targetProdAfter.Heat,
		"Target heat production should decrease by 1")
}

// --- Capital (008) ---
// "Decrease your energy production 2 steps and increase your M€ production 5 steps. Place a city tile."

func TestCapital_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-capital-test",
		Name: "Capital",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 26,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 5, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 3,
	})
	p.Hand().AddCard("card-capital-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 26}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-capital-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Capital should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+5, prodAfter.Credits,
		"Credit production should increase by 5")
	testutil.AssertEqual(t, prodBefore.Energy-2, prodAfter.Energy,
		"Energy production should decrease by 2")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// --- Big Asteroid (011) ---
// "Raise temperature 2 steps and gain 4 titanium. Remove up to 4 plants from any player."

func TestBigAsteroid_TempTitaniumAndRemovePlants(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-big-asteroid-test",
		Name: "Big Asteroid",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 27,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 4, Target: "any-player"},
					{ResourceType: shared.ResourceTitanium, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceTemperature, Amount: 2, Target: "none"},
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
		shared.ResourcePlant: 10,
	})
	attacker.Hand().AddCard("card-big-asteroid-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 27}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-big-asteroid-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Big Asteroid should play successfully")

	attackerResources := attacker.Resources().Get()
	testutil.AssertEqual(t, 4, attackerResources.Titanium, "Attacker should gain 4 titanium")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 6, targetResources.Plants, "Target should have 6 plants (10 - 4)")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+4, tempAfter,
		"Temperature should increase by 2 steps (4 degrees)")
}

// --- Space Elevator (013) ---
// Auto behavior: "Increase your titanium production 1 step."
// Action behavior: "Spend 1 steel to gain 5 M€."

func TestSpaceElevator_TitaniumProductionOnPlay(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-space-elevator-test",
		Name: "Space Elevator",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 27,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 5, Target: "self-player"},
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
	p.Hand().AddCard("card-space-elevator-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 27}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-space-elevator-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Space Elevator should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1")
}

func TestSpaceElevator_ActionSpendSteelGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-space-elevator"
	p.PlayedCards().AddCard(cardID, "Space Elevator", "active", []string{"building", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 3,
	})
	creditsBefore := p.Resources().Get().Credits

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 5, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Space Elevator",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Space Elevator action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Steel, "Should have 2 steel after spending 1")
	testutil.AssertEqual(t, creditsBefore+5, resources.Credits, "Should gain 5 credits")
}

// --- Equatorial Magnetizer (015) ---
// "Action: Decrease your energy production 1 step to increase your terraform rating 1 step."

func TestEquatorialMagnetizer_DecreaseEnergyProdIncreaseTR(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-equatorial-magnetizer"
	p.PlayedCards().AddCard(cardID, "Equatorial Magnetizer", "active", []string{"building"})

	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTR, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Equatorial Magnetizer",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	trBefore := p.Resources().TerraformRating()
	prodBefore := p.Resources().Production()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Equatorial Magnetizer action should succeed")

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
}

// --- Domed Crater (016) ---
// "Gain 3 plants and place a city tile. Decrease your energy production 1 step and increase M€ production 3 steps."

func TestDomedCrater_PlantsProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-domed-crater-test",
		Name: "Domed Crater",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 24,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-domed-crater-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 24}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-domed-crater-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Domed Crater should play successfully")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+3, plantsAfter, "Should gain 3 plants")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// --- Noctis City (017) ---
// "Decrease your energy production 1 step and increase your M€ production 3 steps. Place a city tile on the reserved area."

func TestNoctisCity_ProductionAndReservedCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-noctis-city-test",
		Name: "Noctis City",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 18,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{
						ResourceType: shared.ResourceCityPlacement,
						Amount:       1,
						Target:       "none",
						TileRestrictions: &shared.TileRestrictions{
							BoardTags: []string{"noctis-city"},
						},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-noctis-city-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-noctis-city-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Noctis City should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
}

// --- Methane From Titan (018) ---
// "Increase your heat production 2 steps and your plant production 2 steps."

func TestMethaneFromTitan_HeatAndPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-methane-from-titan-test",
		Name: "Methane From Titan",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 28,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-methane-from-titan-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 28}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-methane-from-titan-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Methane From Titan should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Heat+2, prodAfter.Heat,
		"Heat production should increase by 2")
}

// --- Phobos Space Haven (021) ---
// "Increase your titanium production 1 step and place a city tile on the reserved area."

func TestPhobosSpaceHaven_TitaniumProdAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-phobos-space-haven-test",
		Name: "Phobos Space Haven",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 25,
		Tags: []shared.CardTag{shared.TagCity, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
					{
						ResourceType: shared.ResourceCityPlacement,
						Amount:       1,
						Target:       "none",
						TileRestrictions: &shared.TileRestrictions{
							BoardTags: []string{"phobos-space-haven"},
						},
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
	p.Hand().AddCard("card-phobos-space-haven-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 25}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-phobos-space-haven-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Phobos Space Haven should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1")
}

// --- Black Polar Dust (022) ---
// "Place an ocean tile. Decrease your M€ production 2 steps and increase your heat production 3 steps."

func TestBlackPolarDust_ProductionAndOceanPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-black-polar-dust-test",
		Name: "Black Polar Dust",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 15,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 3,
	})
	p.Hand().AddCard("card-black-polar-dust-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-black-polar-dust-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Black Polar Dust should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-2, prodAfter.Credits,
		"Credit production should decrease by 2")
	testutil.AssertEqual(t, prodBefore.Heat+3, prodAfter.Heat,
		"Heat production should increase by 3")
}

// --- Arctic Algae (023) ---
// Auto: "Gain 1 plant."
// Passive: "When anyone places an ocean tile, gain 2 plants."

func TestArcticAlgae_GainPlantOnPlay(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-arctic-algae-test",
		Name: "Arctic Algae",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-arctic-algae-test")

	plantsBefore := p.Resources().Get().Plants

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-arctic-algae-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Arctic Algae should play successfully")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+1, plantsAfter, "Should gain 1 plant on play")
}

// --- Eos Chasma National Park (026) ---
// "Add 1 animal to any animal card. Gain 3 plants. Increase your M€ production 2 steps."

func TestEosChasmaNationalPark_PlantsAndProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-eos-chasma-test",
		Name: "Eos Chasma National Park",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 16,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "any-card"},
				},
			},
		},
	}

	animalHost := gamecards.Card{
		ID:              "card-animal-host-eos",
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
	p.PlayedCards().AddCard("card-animal-host-eos", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host-eos", 0)
	p.Hand().AddCard("card-eos-chasma-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	targetCardID := "card-animal-host-eos"
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-eos-chasma-test", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Eos Chasma National Park should play successfully")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+3, plantsAfter, "Should gain 3 plants")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits,
		"Credit production should increase by 2")

	animalStorage := p.Resources().GetCardStorage("card-animal-host-eos")
	testutil.AssertEqual(t, 1, animalStorage, "Animal host should have 1 animal")
}

// --- Security Fleet (028) ---
// "Action: Spend 1 titanium to add 1 fighter resource to this card."

func TestSecurityFleet_SpendTitaniumAddFighter(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-security-fleet"
	p.PlayedCards().AddCard(cardID, "Security Fleet", "active", []string{"space"})
	p.Resources().AddToStorage(cardID, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTitanium, Amount: 1, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceFighter, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Security Fleet",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Security Fleet action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Titanium, "Should have 2 titanium after spending 1")

	fighterStorage := p.Resources().GetCardStorage(cardID)
	testutil.AssertEqual(t, 1, fighterStorage, "Security Fleet should have 1 fighter")
}

// --- Cupola City (029) ---
// "Place a city tile. Decrease your energy production 1 step and increase your M€ production 3 steps."

func TestCupolaCity_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-cupola-city-test",
		Name: "Cupola City",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 16,
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-cupola-city-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-cupola-city-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Cupola City should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// --- Lunar Beam (030) ---
// "Decrease your M€ production 2 steps and increase your heat production and energy production 2 steps each."

func TestLunarBeam_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-lunar-beam-test",
		Name: "Lunar Beam",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 3,
	})
	p.Hand().AddCard("card-lunar-beam-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lunar-beam-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lunar Beam should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-2, prodAfter.Credits,
		"Credit production should decrease by 2")
	testutil.AssertEqual(t, prodBefore.Energy+2, prodAfter.Energy,
		"Energy production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Heat+2, prodAfter.Heat,
		"Heat production should increase by 2")
}

// --- Underground City (032) ---
// "Place a city tile. Decrease your energy production 2 steps and increase your steel production 2 steps."

func TestUndergroundCity_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-underground-city-test",
		Name: "Underground City",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 18,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteelProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 3,
	})
	p.Hand().AddCard("card-underground-city-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-underground-city-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Underground City should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+2, prodAfter.Steel,
		"Steel production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Energy-2, prodAfter.Energy,
		"Energy production should decrease by 2")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// --- Release Of Inert Gases (036) ---
// "Raise your terraform rating 2 steps."

func TestReleaseOfInertGases_RaiseTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-release-inert-gases-test",
		Name: "Release Of Inert Gases",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 14,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTR, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-release-inert-gases-test")

	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-release-inert-gases-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Release Of Inert Gases should play successfully")

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+2, trAfter, "TR should increase by 2")
}

// --- Deimos Down (039) ---
// "Raise temperature 3 steps and gain 4 steel. Remove up to 8 plants from any player."

func TestDeimosDown_TempSteelAndRemovePlants(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-deimos-down-test",
		Name: "Deimos Down",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 31,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 8, Target: "any-player"},
					{ResourceType: shared.ResourceSteel, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceTemperature, Amount: 3, Target: "none"},
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
		shared.ResourcePlant: 10,
	})
	attacker.Hand().AddCard("card-deimos-down-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 31}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-deimos-down-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Deimos Down should play successfully")

	attackerResources := attacker.Resources().Get()
	testutil.AssertEqual(t, 4, attackerResources.Steel, "Attacker should gain 4 steel")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Plants, "Target should have 2 plants (10 - 8)")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+6, tempAfter,
		"Temperature should increase by 3 steps (6 degrees)")
}

// --- Asteroid Mining (040) ---
// "Increase your titanium production 2 steps."

func TestAsteroidMining_TitaniumProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-asteroid-mining-test",
		Name: "Asteroid Mining",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 30,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-asteroid-mining-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 30}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-asteroid-mining-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Asteroid Mining should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+2, prodAfter.Titanium,
		"Titanium production should increase by 2")
}

// --- Food Factory (041) ---
// "Decrease your plant production 1 step and increase your M€ production 4 steps."

func TestFoodFactory_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-food-factory-test",
		Name: "Food Factory",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 2,
	})
	p.Hand().AddCard("card-food-factory-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-food-factory-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Food Factory should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+4, prodAfter.Credits,
		"Credit production should increase by 4")
	testutil.AssertEqual(t, prodBefore.Plants-1, prodAfter.Plants,
		"Plant production should decrease by 1")
}

// --- Archaebacteria (042) ---
// "Increase your plant production 1 step."

func TestArchaebacteria_PlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-archaebacteria-test",
		Name: "Archaebacteria",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-archaebacteria-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-archaebacteria-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Archaebacteria should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants,
		"Plant production should increase by 1")
}

// --- Carbonate Processing (043) ---
// "Decrease your energy production 1 step and increase your heat production 3 steps."

func TestCarbonateProcessing_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-carbonate-processing-test",
		Name: "Carbonate Processing",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 3, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-carbonate-processing-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-carbonate-processing-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Carbonate Processing should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
	testutil.AssertEqual(t, prodBefore.Heat+3, prodAfter.Heat,
		"Heat production should increase by 3")
}

// --- Nuclear Power (045) ---
// "Decrease your M€ production 2 steps and increase your energy production 3 steps."

func TestNuclearPower_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-nuclear-power-test",
		Name: "Nuclear Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 3,
	})
	p.Hand().AddCard("card-nuclear-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-nuclear-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuclear Power should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-2, prodAfter.Credits,
		"Credit production should decrease by 2")
	testutil.AssertEqual(t, prodBefore.Energy+3, prodAfter.Energy,
		"Energy production should increase by 3")
}

// --- Lightning Harvest (046) ---
// "Increase your energy production and your M€ production 1 step each."

func TestLightningHarvest_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-lightning-harvest-test",
		Name: "Lightning Harvest",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-lightning-harvest-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lightning-harvest-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lightning Harvest should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits,
		"Credit production should increase by 1")
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1")
}

// --- Algae (047) ---
// "Gain 1 plant and increase your plant production 2 steps."

func TestAlgae_PlantAndPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-algae-test",
		Name: "Algae",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-algae-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-algae-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Algae should play successfully")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+1, plantsAfter, "Should gain 1 plant")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")
}

// --- Adapted Lichen (048) ---
// "Increase your plant production 1 step."

func TestAdaptedLichen_PlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-adapted-lichen-test",
		Name: "Adapted Lichen",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-adapted-lichen-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-adapted-lichen-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Adapted Lichen should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants,
		"Plant production should increase by 1")
}

// --- Tardigrades (049) ---
// "Action: Add 1 microbe to this card."

func TestTardigrades_AddMicrobe(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-tardigrades"
	p.PlayedCards().AddCard(cardID, "Tardigrades", "active", []string{"microbe"})
	p.Resources().AddToStorage(cardID, 0)

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Tardigrades",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Tardigrades add microbe should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID),
		"Tardigrades should have 1 microbe after action")
}

func TestTardigrades_AccumulateMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-tardigrades"
	p.PlayedCards().AddCard(cardID, "Tardigrades", "active", []string{"microbe"})
	p.Resources().AddToStorage(cardID, 3) // Start with 3 microbes

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Tardigrades",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Tardigrades add microbe should succeed")

	testutil.AssertEqual(t, 4, p.Resources().GetCardStorage(cardID),
		"Tardigrades should have 4 microbes (3 + 1)")
}

// --- Fish (052) ---
// Auto: "Decrease any plant production 1 step."
// Action: "Add 1 animal to this card."

func TestFish_DecreasePlantProductionOnPlay(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-fish-test",
		Name: "Fish",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagAnimal},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "any-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
				},
			},
		},
		ResourceStorage: &gamecards.ResourceStorage{
			Type:     shared.ResourceAnimal,
			Starting: 0,
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
	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 3,
	})
	attacker.Hand().AddCard("card-fish-test")

	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-fish-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Fish should play successfully")

	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, targetProdBefore.Plants-1, targetProdAfter.Plants,
		"Target plant production should decrease by 1")
}

func TestFish_ActionAddAnimal(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-fish"
	p.PlayedCards().AddCard(cardID, "Fish", "active", []string{"animal"})
	p.Resources().AddToStorage(cardID, 0)

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Fish",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Fish add animal action should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID),
		"Fish should have 1 animal after action")
}

// --- Comet (010) ---
// "Raise temperature 1 step and place an ocean tile. Remove up to 3 plants from any player."

func TestComet_TempOceanAndRemovePlants(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-comet-test",
		Name: "Comet",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 21,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "any-player"},
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
		shared.ResourcePlant: 5,
	})
	attacker.Hand().AddCard("card-comet-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-comet-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Comet should play successfully")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+2, tempAfter,
		"Temperature should increase by 1 step (2 degrees)")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Plants, "Target should have 2 plants (5 - 3)")
}

// --- Asteroid (009) ---
// "Raise temperature 1 step and gain 2 titanium. Remove up to 3 plants from any player."
// (Testing the self-player titanium gain specifically)

func TestAsteroid_SelfPlayerGainTitanium(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-asteroid-solo-test",
		Name: "Asteroid",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 14,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "any-player"},
					{ResourceType: shared.ResourceTitanium, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "none"},
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
	p.Hand().AddCard("card-asteroid-solo-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-asteroid-solo-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Asteroid should play successfully in solo mode")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Titanium, "Should gain 2 titanium")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+2, tempAfter,
		"Temperature should increase by 1 step (2 degrees)")
}

// --- Interstellar Colony Ship (027) ---
// No behaviors - just 4 VP. Requires 5 science tags.

func TestInterstellarColonyShip_RequiresScienteTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	sciTag := shared.TagScience
	card := gamecards.Card{
		ID:   "card-interstellar-colony-ship-test",
		Name: "Interstellar Colony Ship",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 24,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(5), Tag: &sciTag},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	// Register science helper cards in the registry so tag counting works
	for i := 0; i < 5; i++ {
		scienceCard := gamecards.Card{
			ID:   "card-science-" + string(rune('a'+i)),
			Name: "Science Card",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 0,
			Tags: []shared.CardTag{shared.TagScience},
		}
		additionalCards = append(additionalCards, scienceCard)
	}
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
	p.Hand().AddCard("card-interstellar-colony-ship-test")

	// Without 5 science tags, should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 24}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-interstellar-colony-ship-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail without 5 science tags")

	// Add 5 science-tagged played cards
	for i := 0; i < 5; i++ {
		cardID := "card-science-" + string(rune('a'+i))
		p.PlayedCards().AddCard(cardID, "Science Card", "automated", []string{"science"})
	}

	// Re-add card to hand since failed play should not remove it
	if !p.Hand().HasCard("card-interstellar-colony-ship-test") {
		p.Hand().AddCard("card-interstellar-colony-ship-test")
	}

	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-interstellar-colony-ship-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should succeed with 5 science tags")
}

// --- Comet (010) Solo Mode ---
// Verify that in solo mode, the any-player plant removal is skipped.

func TestComet_SoloMode_PlantRemovalSkipped(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-comet-solo-test",
		Name: "Comet",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 21,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "any-player"},
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
		shared.ResourcePlant:  10,
	})
	p.Hand().AddCard("card-comet-solo-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-comet-solo-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Comet should play successfully in solo mode")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 10, resources.Plants,
		"Player's plants should be unchanged in solo mode (any-player removal skipped)")
}

// --- Deimos Down (039) Solo Mode ---
// In solo mode, plant removal from any-player is skipped but steel gain and temperature work.

func TestDeimosDown_SoloMode_SteelAndTemp(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-deimos-down-solo-test",
		Name: "Deimos Down",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 31,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 8, Target: "any-player"},
					{ResourceType: shared.ResourceSteel, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceTemperature, Amount: 3, Target: "none"},
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
		shared.ResourcePlant:  10,
	})
	p.Hand().AddCard("card-deimos-down-solo-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 31}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-deimos-down-solo-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Deimos Down should play in solo mode")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 4, resources.Steel, "Should gain 4 steel")
	testutil.AssertEqual(t, 10, resources.Plants,
		"Plants should be unchanged in solo mode (any-player removal skipped)")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+6, tempAfter,
		"Temperature should increase by 3 steps (6 degrees)")
}

// --- Methane From Titan (018) - with oxygen requirement ---
// This card requires 2% oxygen. Test that it fails when oxygen is too low.

func TestMethaneFromTitan_FailsWithoutOxygenRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-methane-titan-req-test",
		Name: "Methane From Titan",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 28,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(2)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-methane-titan-req-test")

	// Oxygen starts at 0%, requirement is min 2%
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 28}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-methane-titan-req-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Methane From Titan should fail at 0% oxygen (requires 2%)")
}

// --- Food Factory (041) - fails without plant production ---
// "Decrease your plant production 1 step" - should fail if player has 0 plant production.

func TestFoodFactory_FailsWithoutPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	plantRes := shared.ResourcePlant
	card := gamecards.Card{
		ID:   "card-food-factory-fail-test",
		Name: "Food Factory",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementProduction, Min: testutil.IntPtr(1), Resource: &plantRes},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "self-player"},
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
	// No plant production - should fail
	p.Hand().AddCard("card-food-factory-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-food-factory-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Food Factory should fail without plant production")
}

// --- Nuclear Power (045) - fails without enough credit production ---
// "Decrease your M€ production 2 steps" - should fail if player has < 2 credit production (below -5 floor).

func TestNuclearPower_FailsWithInsufficientCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	creditRes := shared.ResourceCredit
	card := gamecards.Card{
		ID:   "card-nuclear-power-fail-test",
		Name: "Nuclear Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementProduction, Min: testutil.IntPtr(-3), Resource: &creditRes},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
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
	// Set credit production to -4 (near -5 floor). Decreasing by 2 would go to -6, below floor.
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: -4,
	})
	p.Hand().AddCard("card-nuclear-power-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-nuclear-power-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err,
		"Nuclear Power should fail when credit production would go below -5")
}

// --- Carbonate Processing (043) - fails without energy production ---
// "Decrease your energy production 1 step" - should fail if player has 0 energy production.

func TestCarbonateProcessing_FailsWithoutEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	energyRes := shared.ResourceEnergy
	card := gamecards.Card{
		ID:   "card-carbonate-fail-test",
		Name: "Carbonate Processing",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementProduction, Min: testutil.IntPtr(1), Resource: &energyRes},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 3, Target: "self-player"},
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
	// No energy production
	p.Hand().AddCard("card-carbonate-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-carbonate-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Carbonate Processing should fail without energy production")
}

// --- Colonizer Training Camp (001) - max oxygen requirement failure ---
// Should fail when oxygen is above 5%.

func TestColonizerTrainingCamp_FailsAboveMaxOxygen(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-colonizer-camp-fail-test",
		Name: "Colonizer Training Camp",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagJovian},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Max: testutil.IntPtr(5)},
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

	// Increase oxygen above 5%
	for i := 0; i < 6; i++ {
		testGame.GlobalParameters().IncreaseOxygen(ctx, 1)
	}

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-colonizer-camp-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-colonizer-camp-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Colonizer Training Camp should fail when oxygen is above 5%")
}

// tagP returns a pointer to a CardTag (unique name to avoid redefinition conflicts).

// iPtr returns a pointer to an int (unique name to avoid redefinition conflicts).

// --- Lake Marineris (053) ---
// "Place 2 ocean tiles. Requires 0°C or warmer."

func TestLakeMarineris_PlacesTwoOceanTiles(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	lakeMarineris := gamecards.Card{
		ID:   "card-lake-marineris",
		Name: "Lake Marineris",
		Type: gamecards.CardTypeAutomated,
		Cost: 18,
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(0)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 2, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{lakeMarineris})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Meet temperature requirement
	testGame.GlobalParameters().SetTemperature(ctx, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-lake-marineris")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lake-marineris", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lake Marineris should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for ocean placement")
	testutil.AssertEqual(t, "ocean", selection.TileType, "Pending tile type should be ocean")
}

func TestLakeMarineris_FailsWithoutTemperatureRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	lakeMarineris := gamecards.Card{
		ID:   "card-lake-marineris",
		Name: "Lake Marineris",
		Type: gamecards.CardTypeAutomated,
		Cost: 18,
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(0)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 2, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{lakeMarineris})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Temperature below requirement (default -30, need 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-lake-marineris")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lake-marineris", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Lake Marineris should fail without meeting temperature requirement")
}

// --- Small Animals (054) ---
// Active card: "Action: Add 1 animal to this card. Decrease any plant production 1 step. Requires 6% oxygen."

func TestSmallAnimals_AddAnimalAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-small-animals"
	p.PlayedCards().AddCard(cardID, "Small Animals", "active", []string{"animal"})
	p.Resources().AddToStorage(cardID, 0)

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Small Animals",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Small Animals add animal action should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 animal after action")
}

// --- Kelp Farming (055) ---
// "Increase your M€ production 2 steps and your plant production 3 steps. Gain 2 plants. Requires 6 oceans."

func TestKelpFarming_ProductionAndPlantGain(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	kelpFarming := gamecards.Card{
		ID:   "card-kelp-farming",
		Name: "Kelp Farming",
		Type: gamecards.CardTypeAutomated,
		Cost: 17,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{kelpFarming})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-kelp-farming")

	plantsBefore := p.Resources().Get().Plants
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-kelp-farming", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Kelp Farming should play successfully")

	plantsAfter := p.Resources().Get().Plants
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")
	testutil.AssertEqual(t, productionBefore.Credits+2, productionAfter.Credits, "Should gain 2 credit production")
	testutil.AssertEqual(t, productionBefore.Plants+3, productionAfter.Plants, "Should gain 3 plant production")
}

// --- Mine (056) ---
// "Increase your steel production 1 step."

func TestMine_SteelProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	mine := gamecards.Card{
		ID:   "card-mine",
		Name: "Mine",
		Type: gamecards.CardTypeAutomated,
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mine})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mine")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mine", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mine should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Steel+1, productionAfter.Steel, "Should gain 1 steel production")
}

// --- Vesta Shipyard (057) ---
// "Increase your titanium production 1 step."

func TestVestaShipyard_TitaniumProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	vestaShipyard := gamecards.Card{
		ID:   "card-vesta-shipyard",
		Name: "Vesta Shipyard",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{vestaShipyard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-vesta-shipyard")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-vesta-shipyard", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Vesta Shipyard should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Titanium+1, productionAfter.Titanium, "Should gain 1 titanium production")
}

// --- Beam From A Thorium Asteroid (058) ---
// "Increase your heat production and energy production 3 steps each. Requires jovian tag."

func TestBeamFromAThoriumAsteroid_HeatAndEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	beamFromAsteroid := gamecards.Card{
		ID:   "card-beam-from-asteroid",
		Name: "Beam From A Thorium Asteroid",
		Type: gamecards.CardTypeAutomated,
		Cost: 32,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagPower, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{beamFromAsteroid})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-beam-from-asteroid")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 32}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-beam-from-asteroid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Beam From A Thorium Asteroid should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Heat+3, productionAfter.Heat, "Should gain 3 heat production")
	testutil.AssertEqual(t, productionBefore.Energy+3, productionAfter.Energy, "Should gain 3 energy production")
}

// --- Trees (060) ---
// "Requires -4°C or warmer. Increase your plant production 3 steps. Gain 1 plant."

func TestTrees_PlantProductionAndGain(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	trees := gamecards.Card{
		ID:   "card-trees",
		Name: "Trees",
		Type: gamecards.CardTypeAutomated,
		Cost: 13,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-4)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{trees})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetTemperature(ctx, -4)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-trees")

	plantsBefore := p.Resources().Get().Plants
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-trees", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Trees should play successfully")

	plantsAfter := p.Resources().Get().Plants
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, plantsBefore+1, plantsAfter, "Should gain 1 plant")
	testutil.AssertEqual(t, productionBefore.Plants+3, productionAfter.Plants, "Should gain 3 plant production")
}

func TestTrees_FailsWithoutTemperatureRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	trees := gamecards.Card{
		ID:   "card-trees",
		Name: "Trees",
		Type: gamecards.CardTypeAutomated,
		Cost: 13,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-4)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{trees})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Temperature below requirement (default -30, need -4)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-trees")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-trees", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Trees should fail without meeting temperature requirement")
}

// --- Mineral Deposit (062) ---
// "Gain 5 steel."

func TestMineralDeposit_GainSteel(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	mineralDeposit := gamecards.Card{
		ID:   "card-mineral-deposit",
		Name: "Mineral Deposit",
		Type: gamecards.CardTypeEvent,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteel, Amount: 5, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mineralDeposit})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mineral-deposit")

	steelBefore := p.Resources().Get().Steel

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mineral-deposit", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mineral Deposit should play successfully")

	steelAfter := p.Resources().Get().Steel
	testutil.AssertEqual(t, steelBefore+5, steelAfter, "Should gain 5 steel")
}

// --- Mining Expedition (063) ---
// "Raise oxygen 1 step. Remove 2 plants from any player. Gain 2 steel."

func TestMiningExpedition_OxygenAndSteelAndRemovePlants(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	miningExpedition := gamecards.Card{
		ID:   "card-mining-expedition",
		Name: "Mining Expedition",
		Type: gamecards.CardTypeEvent,
		Cost: 12,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "any-player"},
					{ResourceType: shared.ResourceSteel, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{miningExpedition})

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
	attacker.Hand().AddCard("card-mining-expedition")

	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 5,
	})

	steelBefore := attacker.Resources().Get().Steel
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-mining-expedition", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Mining Expedition should play successfully")

	steelAfter := attacker.Resources().Get().Steel
	testutil.AssertEqual(t, steelBefore+2, steelAfter, "Should gain 2 steel")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")

	targetPlants := target.Resources().Get().Plants
	testutil.AssertEqual(t, 3, targetPlants, "Target should have 3 plants after removing 2")
}

// --- Building Industries (065) ---
// "Decrease your energy production 1 step and increase your steel production 2 steps."

func TestBuildingIndustries_EnergyToSteelProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	buildingIndustries := gamecards.Card{
		ID:   "card-building-industries",
		Name: "Building Industries",
		Type: gamecards.CardTypeAutomated,
		Cost: 6,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceSteelProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{buildingIndustries})

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
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-building-industries")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-building-industries", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Building Industries should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Energy-1, productionAfter.Energy, "Energy production should decrease by 1")
	testutil.AssertEqual(t, productionBefore.Steel+2, productionAfter.Steel, "Steel production should increase by 2")
}

// --- Sponsors (068) ---
// "Increase your M€ production 2 steps."

func TestSponsors_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	sponsors := gamecards.Card{
		ID:   "card-sponsors",
		Name: "Sponsors",
		Type: gamecards.CardTypeAutomated,
		Cost: 6,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{sponsors})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-sponsors")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-sponsors", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Sponsors should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+2, productionAfter.Credits, "Should gain 2 credit production")
}

// --- Towing A Comet (075) ---
// "Gain 2 plants. Raise oxygen level 1 step and place an ocean tile."

func TestTowingAComet_PlantsOxygenOcean(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	towingAComet := gamecards.Card{
		ID:   "card-towing-a-comet",
		Name: "Towing A Comet",
		Type: gamecards.CardTypeEvent,
		Cost: 23,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{towingAComet})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-towing-a-comet")

	plantsBefore := p.Resources().Get().Plants
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-towing-a-comet", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Towing A Comet should play successfully")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")
}

// --- Space Mirrors (076) ---
// Active card: "Action: Spend 7 M€ to increase your energy production 1 step."

func TestSpaceMirrors_SpendCreditsForEnergyProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-space-mirrors"
	p.PlayedCards().AddCard(cardID, "Space Mirrors", "active", []string{"power", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Space Mirrors",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	productionBefore := p.Resources().Production()
	creditsBefore := p.Resources().Get().Credits

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Space Mirrors action should succeed")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Energy+1, productionAfter.Energy, "Energy production should increase by 1")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-7, creditsAfter, "Should spend 7 credits")
}

// --- Solar Wind Power (077) ---
// "Increase your energy production 1 step and gain 2 titanium."

func TestSolarWindPower_EnergyProductionAndTitanium(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	solarWindPower := gamecards.Card{
		ID:   "card-solar-wind-power",
		Name: "Solar Wind Power",
		Type: gamecards.CardTypeAutomated,
		Cost: 11,
		Tags: []shared.CardTag{shared.TagPower, shared.TagScience, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitanium, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{solarWindPower})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-solar-wind-power")

	titaniumBefore := p.Resources().Get().Titanium
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-solar-wind-power", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Solar Wind Power should play successfully")

	titaniumAfter := p.Resources().Get().Titanium
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, titaniumBefore+2, titaniumAfter, "Should gain 2 titanium")
	testutil.AssertEqual(t, productionBefore.Energy+1, productionAfter.Energy, "Should gain 1 energy production")
}

// --- Ice Asteroid (078) ---
// "Place 2 ocean tiles."

func TestIceAsteroid_PlaceTwoOceanTiles(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	iceAsteroid := gamecards.Card{
		ID:   "card-ice-asteroid",
		Name: "Ice Asteroid",
		Type: gamecards.CardTypeEvent,
		Cost: 23,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 2, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{iceAsteroid})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ice-asteroid")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ice-asteroid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ice Asteroid should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for ocean placement")
	testutil.AssertEqual(t, "ocean", selection.TileType, "Pending tile type should be ocean")
}

// --- Callisto Penal Mines (082) ---
// "Increase your M€ production 3 steps."

func TestCallistoPenalMines_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	callistoPenalMines := gamecards.Card{
		ID:   "card-callisto-penal-mines",
		Name: "Callisto Penal Mines",
		Type: gamecards.CardTypeAutomated,
		Cost: 24,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{callistoPenalMines})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-callisto-penal-mines")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 24}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-callisto-penal-mines", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Callisto Penal Mines should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+3, productionAfter.Credits, "Should gain 3 credit production")
}

// --- Giant Space Mirror (083) ---
// "Increase your energy production 3 steps."

func TestGiantSpaceMirror_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	giantSpaceMirror := gamecards.Card{
		ID:   "card-giant-space-mirror",
		Name: "Giant Space Mirror",
		Type: gamecards.CardTypeAutomated,
		Cost: 17,
		Tags: []shared.CardTag{shared.TagPower, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{giantSpaceMirror})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-giant-space-mirror")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-giant-space-mirror", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Giant Space Mirror should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Energy+3, productionAfter.Energy, "Should gain 3 energy production")
}

// --- Commercial District (085) ---
// "Decrease your energy production 1 step and increase your M€ production 4 steps."

func TestCommercialDistrict_ProductionChange(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	commercialDistrict := gamecards.Card{
		ID:   "card-commercial-district",
		Name: "Commercial District",
		Type: gamecards.CardTypeAutomated,
		Cost: 16,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{commercialDistrict})

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
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-commercial-district")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-commercial-district", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Commercial District should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+4, productionAfter.Credits, "Should gain 4 credit production")
	testutil.AssertEqual(t, productionBefore.Energy-1, productionAfter.Energy, "Energy production should decrease by 1")
}

// --- Grass (087) ---
// "Requires -16°C or warmer. Increase your plant production 1 step. Gain 3 plants."

func TestGrass_PlantProductionAndGain(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	grass := gamecards.Card{
		ID:   "card-grass",
		Name: "Grass",
		Type: gamecards.CardTypeAutomated,
		Cost: 11,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-16)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{grass})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetTemperature(ctx, -16)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-grass")

	plantsBefore := p.Resources().Get().Plants
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-grass", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Grass should play successfully")

	plantsAfter := p.Resources().Get().Plants
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, plantsBefore+3, plantsAfter, "Should gain 3 plants")
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants, "Should gain 1 plant production")
}

// --- Heather (088) ---
// "Requires -14°C or warmer. Increase your plant production 1 step. Gain 1 plant."

func TestHeather_PlantProductionAndGain(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	heather := gamecards.Card{
		ID:   "card-heather",
		Name: "Heather",
		Type: gamecards.CardTypeAutomated,
		Cost: 6,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-14)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{heather})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetTemperature(ctx, -14)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-heather")

	plantsBefore := p.Resources().Get().Plants
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-heather", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Heather should play successfully")

	plantsAfter := p.Resources().Get().Plants
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, plantsBefore+1, plantsAfter, "Should gain 1 plant")
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants, "Should gain 1 plant production")
}

func TestHeather_FailsWithoutTemperature(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	heather := gamecards.Card{
		ID:   "card-heather",
		Name: "Heather",
		Type: gamecards.CardTypeAutomated,
		Cost: 6,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-14)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{heather})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Default temperature is -30, need -14

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-heather")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-heather", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Heather should fail without meeting temperature requirement")
}

// --- Peroxide Power (089) ---
// "Decrease your M€ production 1 step and increase your energy production 2 steps."

func TestPeroxidePower_ProductionChange(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	peroxidePower := gamecards.Card{
		ID:   "card-peroxide-power",
		Name: "Peroxide Power",
		Type: gamecards.CardTypeAutomated,
		Cost: 7,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{peroxidePower})

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
		shared.ResourceCreditProduction: 1,
	})
	p.Hand().AddCard("card-peroxide-power")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-peroxide-power", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Peroxide Power should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits-1, productionAfter.Credits, "Credit production should decrease by 1")
	testutil.AssertEqual(t, productionBefore.Energy+2, productionAfter.Energy, "Energy production should increase by 2")
}

// --- Research (090) ---
// "Counts as playing 2 science cards. Draw 2 cards."

func TestResearch_DrawTwoCards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	research := gamecards.Card{
		ID:   "card-research",
		Name: "Research",
		Type: gamecards.CardTypeAutomated,
		Cost: 11,
		Tags: []shared.CardTag{shared.TagScience, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{research})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-research")

	handBefore := p.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-research", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Research should play successfully")

	// Played 1 card (-1), drew 2 cards (+2) = net +1
	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore+1, handAfter, "Hand should increase by 1 (played 1, drew 2)")
}

// --- Gene Repair (091) ---
// "Requires 3 science tags. Increase your M€ production 2 steps."

func TestGeneRepair_CreditProductionWithScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	scienceTag := shared.TagScience
	geneRepair := gamecards.Card{
		ID:   "card-gene-repair",
		Name: "Gene Repair",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTags, Min: testutil.IntPtr(3), Tag: &scienceTag},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	sciCard1 := gamecards.Card{ID: "card-sci-1", Name: "Science Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard2 := gamecards.Card{ID: "card-sci-2", Name: "Science Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{geneRepair, sciCard1, sciCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Give player 2 existing science tags (gene repair itself has 1, making it 3 total)
	p.PlayedCards().AddCard("card-sci-1", "Science Card 1", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-2", "Science Card 2", "automated", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-gene-repair")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-gene-repair", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Gene Repair should play successfully with 3 science tags")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+2, productionAfter.Credits, "Should gain 2 credit production")
}

func TestGeneRepair_FailsWithoutScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	scienceTag := shared.TagScience
	geneRepair := gamecards.Card{
		ID:   "card-gene-repair",
		Name: "Gene Repair",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTags, Min: testutil.IntPtr(3), Tag: &scienceTag},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{geneRepair})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Only 1 science tag (from gene repair itself) -- not enough
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-gene-repair")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-gene-repair", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Gene Repair should fail without 3 science tags")
}

// --- Io Mining Industries (092) ---
// "Increase your titanium production 2 steps and your M€ production 2 steps."

func TestIoMiningIndustries_TitaniumAndCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ioMining := gamecards.Card{
		ID:   "card-io-mining",
		Name: "Io Mining Industries",
		Type: gamecards.CardTypeAutomated,
		Cost: 41,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ioMining})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-io-mining")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 41}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-io-mining", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Io Mining Industries should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+2, productionAfter.Credits, "Should gain 2 credit production")
	testutil.AssertEqual(t, productionBefore.Titanium+2, productionAfter.Titanium, "Should gain 2 titanium production")
}

// --- Bushes (093) ---
// "Requires -10°C or warmer. Increase your plant production 2 steps. Gain 2 plants."

func TestBushes_PlantProductionAndGain(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	bushes := gamecards.Card{
		ID:   "card-bushes",
		Name: "Bushes",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-10)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{bushes})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetTemperature(ctx, -10)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-bushes")

	plantsBefore := p.Resources().Get().Plants
	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-bushes", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Bushes should play successfully")

	plantsAfter := p.Resources().Get().Plants
	productionAfter := p.Resources().Production()

	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")
	testutil.AssertEqual(t, productionBefore.Plants+2, productionAfter.Plants, "Should gain 2 plant production")
}

// --- Physics Complex (095) ---
// Active card: "Action: Spend 6 energy to add a science resource to this card."

func TestPhysicsComplex_SpendEnergyForScience(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-physics-complex"
	p.PlayedCards().AddCard(cardID, "Physics Complex", "active", []string{"building", "science"})
	p.Resources().AddToStorage(cardID, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 6, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Physics Complex",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Physics Complex action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 4, resources.Energy, "Should have 4 energy after spending 6")

	storage := p.Resources().GetCardStorage(cardID)
	testutil.AssertEqual(t, 1, storage, "Card should have 1 science resource")
}

func TestPhysicsComplex_FailsWithoutEnoughEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-physics-complex"
	p.PlayedCards().AddCard(cardID, "Physics Complex", "active", []string{"building", "science"})
	p.Resources().AddToStorage(cardID, 0)

	// Only 3 energy, need 6
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 6, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceScience, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Physics Complex",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Physics Complex should fail without enough energy")
}

// --- Tropical Resort (098) ---
// "Decrease your heat production 2 steps and increase your M€ production 3 steps."

func TestTropicalResort_ProductionChange(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	tropicalResort := gamecards.Card{
		ID:   "card-tropical-resort",
		Name: "Tropical Resort",
		Type: gamecards.CardTypeAutomated,
		Cost: 13,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: -2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{tropicalResort})

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
		shared.ResourceHeatProduction: 2,
	})
	p.Hand().AddCard("card-tropical-resort")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-tropical-resort", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Tropical Resort should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+3, productionAfter.Credits, "Should gain 3 credit production")
	testutil.AssertEqual(t, productionBefore.Heat-2, productionAfter.Heat, "Heat production should decrease by 2")
}

// --- Toll Station (099) ---
// "Increase your M€ production 1 step for each space tag your opponents have."

func TestTollStation_CreditProductionPerOpponentSpaceTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	otherPlayersTarget := "other-players"
	spaceTag := shared.TagSpace

	tollStation := gamecards.Card{
		ID:   "card-toll-station",
		Name: "Toll Station",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: "tag",
							Amount:       1,
							Location:     testutil.StrPtr("anywhere"),
							Target:       &otherPlayersTarget,
							Tag:          &spaceTag,
						},
					},
				},
			},
		},
	}

	spaceCard1 := gamecards.Card{ID: "card-space-1", Name: "Space Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagSpace}}
	spaceCard2 := gamecards.Card{ID: "card-space-2", Name: "Space Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagSpace}}
	spaceCard3 := gamecards.Card{ID: "card-space-3", Name: "Space Card 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagSpace}}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{tollStation, spaceCard1, spaceCard2, spaceCard3})

	players := testGame.GetAllPlayers()
	attacker := players[0]
	opponent := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	opponent.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	// Opponent has 3 space tags
	opponent.PlayedCards().AddCard("card-space-1", "Space Card 1", "automated", []string{"space"})
	opponent.PlayedCards().AddCard("card-space-2", "Space Card 2", "automated", []string{"space"})
	opponent.PlayedCards().AddCard("card-space-3", "Space Card 3", "automated", []string{"space"})

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-toll-station")

	productionBefore := attacker.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-toll-station", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Toll Station should play successfully")

	productionAfter := attacker.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+3, productionAfter.Credits,
		"Should gain 3 credit production (1 per each of 3 opponent space tags)")
}

// strPtrB2 returns a pointer to a string (unique name to avoid redefinition conflicts).

// --- Fueled Generators (100) ---
// "Decrease your M€ production 1 step and increase your energy production 1 step."

func TestFueledGenerators_ProductionChange(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	fueledGenerators := gamecards.Card{
		ID:   "card-fueled-generators",
		Name: "Fueled Generators",
		Type: gamecards.CardTypeAutomated,
		Cost: 1,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{fueledGenerators})

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
		shared.ResourceCreditProduction: 1,
	})
	p.Hand().AddCard("card-fueled-generators")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-fueled-generators", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Fueled Generators should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits-1, productionAfter.Credits, "Credit production should decrease by 1")
	testutil.AssertEqual(t, productionBefore.Energy+1, productionAfter.Energy, "Energy production should increase by 1")
}

// --- Ironworks (101) ---
// Active card: "Action: Spend 4 energy to gain 1 steel and increase oxygen 1 step."

func TestIronworks_SpendEnergyForSteelAndOxygen(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-ironworks"
	p.PlayedCards().AddCard(cardID, "Ironworks", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 6,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Ironworks",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	steelBefore := p.Resources().Get().Steel
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ironworks action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Energy, "Should have 2 energy after spending 4")
	testutil.AssertEqual(t, steelBefore+1, resources.Steel, "Should gain 1 steel")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")
}

func TestIronworks_FailsWithoutEnoughEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-ironworks"
	p.PlayedCards().AddCard(cardID, "Ironworks", "active", []string{"building"})

	// Only 2 energy, need 4
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 2,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"},
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Ironworks",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Ironworks should fail without enough energy")
}

// --- Power Grid (102) ---
// "Increase your energy production 1 step for each power tag you have, including this."

func TestPowerGrid_EnergyProductionPerPowerTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	powerTag := shared.TagPower

	powerGrid := gamecards.Card{
		ID:   "card-power-grid",
		Name: "Power Grid",
		Type: gamecards.CardTypeAutomated,
		Cost: 18,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceEnergyProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: "tag",
							Amount:       1,
							Location:     testutil.StrPtr("anywhere"),
							Target:       &selfPlayerTarget,
							Tag:          &powerTag,
						},
					},
				},
			},
		},
	}

	powerCard1 := gamecards.Card{ID: "card-power-1", Name: "Power Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagPower}}
	powerCard2 := gamecards.Card{ID: "card-power-2", Name: "Power Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagPower}}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{powerGrid, powerCard1, powerCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Give player 2 existing power tags
	p.PlayedCards().AddCard("card-power-1", "Power Card 1", "automated", []string{"power"})
	p.PlayedCards().AddCard("card-power-2", "Power Card 2", "automated", []string{"power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-power-grid")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-power-grid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Power Grid should play successfully")

	productionAfter := p.Resources().Production()
	// 2 existing power tags + 1 from Power Grid itself = 3
	testutil.AssertEqual(t, productionBefore.Energy+3, productionAfter.Energy,
		"Should gain 3 energy production (1 per each of 2 existing + 1 self power tag)")
}

func TestPowerGrid_NoPreviousPowerTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	powerTag := shared.TagPower

	powerGrid := gamecards.Card{
		ID:   "card-power-grid",
		Name: "Power Grid",
		Type: gamecards.CardTypeAutomated,
		Cost: 18,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceEnergyProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: "tag",
							Amount:       1,
							Location:     testutil.StrPtr("anywhere"),
							Target:       &selfPlayerTarget,
							Tag:          &powerTag,
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{powerGrid})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// No power tags except from Power Grid itself

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-power-grid")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-power-grid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Power Grid should play successfully with no prior power tags")

	productionAfter := p.Resources().Production()
	// Only Power Grid itself = 1 power tag
	testutil.AssertEqual(t, productionBefore.Energy+1, productionAfter.Energy,
		"Should gain 1 energy production (only self power tag)")
}

// --- Steelworks (103) ---
// Active card: "Action: Spend 4 energy to gain 2 steel and increase oxygen 1 step."

func TestSteelworks_SpendEnergyForSteelAndOxygen(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-steelworks"
	p.PlayedCards().AddCard(cardID, "Steelworks", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 8,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteel, Amount: 2, Target: "self-player"},
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Steelworks",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	steelBefore := p.Resources().Get().Steel
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Steelworks action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 4, resources.Energy, "Should have 4 energy after spending 4")
	testutil.AssertEqual(t, steelBefore+2, resources.Steel, "Should gain 2 steel")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")
}

// --- Ore Processor (104) ---
// Active card: "Action: Spend 4 energy to gain 1 titanium and increase oxygen 1 step."

func TestOreProcessor_SpendEnergyForTitaniumAndOxygen(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-ore-processor"
	p.PlayedCards().AddCard(cardID, "Ore Processor", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 5,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTitanium, Amount: 1, Target: "self-player"},
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Ore Processor",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	titaniumBefore := p.Resources().Get().Titanium
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ore Processor action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 1, resources.Energy, "Should have 1 energy after spending 4")
	testutil.AssertEqual(t, titaniumBefore+1, resources.Titanium, "Should gain 1 titanium")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")
}

func TestOreProcessor_FailsWithoutEnoughEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-ore-processor"
	p.PlayedCards().AddCard(cardID, "Ore Processor", "active", []string{"building"})

	// Only 3 energy, need 4
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTitanium, Amount: 1, Target: "self-player"},
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Ore Processor",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Ore Processor should fail without enough energy")
}

// --- Mass Converter (094) ---
// "Requires 5 science tags. Increase your energy production 6 steps."

func TestMassConverter_EnergyProductionWithScienceRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	scienceTag := shared.TagScience
	massConverter := gamecards.Card{
		ID:   "card-mass-converter",
		Name: "Mass Converter",
		Type: gamecards.CardTypeActive,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagPower, shared.TagScience},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTags, Min: testutil.IntPtr(5), Tag: &scienceTag},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player",
						Selectors: []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}}},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 6, Target: "self-player"},
				},
			},
		},
	}

	sciCard1 := gamecards.Card{ID: "card-sci-1", Name: "Science Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard2 := gamecards.Card{ID: "card-sci-2", Name: "Science Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard3 := gamecards.Card{ID: "card-sci-3", Name: "Science Card 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard4 := gamecards.Card{ID: "card-sci-4", Name: "Science Card 4", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{massConverter, sciCard1, sciCard2, sciCard3, sciCard4})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Give player 4 existing science tags (Mass Converter itself has 1, making it 5 total)
	p.PlayedCards().AddCard("card-sci-1", "Science Card 1", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-2", "Science Card 2", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-3", "Science Card 3", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-4", "Science Card 4", "automated", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mass-converter")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mass-converter", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mass Converter should play successfully with 5 science tags")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Energy+6, productionAfter.Energy, "Should gain 6 energy production")
}

func TestMassConverter_FailsWithoutScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	scienceTag := shared.TagScience
	massConverter := gamecards.Card{
		ID:   "card-mass-converter",
		Name: "Mass Converter",
		Type: gamecards.CardTypeActive,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagPower, shared.TagScience},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTags, Min: testutil.IntPtr(5), Tag: &scienceTag},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 6, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{massConverter})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Only 1 science tag from Mass Converter itself, need 5
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mass-converter")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mass-converter", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Mass Converter should fail without 5 science tags")
}

// --- Quantum Extractor (079) ---
// "Requires 4 science tags. Increase your energy production 4 steps."

func TestQuantumExtractor_EnergyProductionWithScienceRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	scienceTag := shared.TagScience
	quantumExtractor := gamecards.Card{
		ID:   "card-quantum-extractor",
		Name: "Quantum Extractor",
		Type: gamecards.CardTypeActive,
		Cost: 13,
		Tags: []shared.CardTag{shared.TagPower, shared.TagScience},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTags, Min: testutil.IntPtr(4), Tag: &scienceTag},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player",
						Selectors: []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}}},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 4, Target: "self-player"},
				},
			},
		},
	}

	sciCard1 := gamecards.Card{ID: "card-sci-1", Name: "Science Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard2 := gamecards.Card{ID: "card-sci-2", Name: "Science Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	sciCard3 := gamecards.Card{ID: "card-sci-3", Name: "Science Card 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{quantumExtractor, sciCard1, sciCard2, sciCard3})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Give player 3 existing science tags (Quantum Extractor has 1, total = 4)
	p.PlayedCards().AddCard("card-sci-1", "Science Card 1", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-2", "Science Card 2", "automated", []string{"science"})
	p.PlayedCards().AddCard("card-sci-3", "Science Card 3", "automated", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-quantum-extractor")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-quantum-extractor", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Quantum Extractor should play successfully with 4 science tags")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Energy+4, productionAfter.Energy, "Should gain 4 energy production")
}

// --- Giant Ice Asteroid (080) ---
// "Raise temperature 2 steps and place 2 ocean tiles. Remove up to 6 plants from any player."

func TestGiantIceAsteroid_TemperatureAndOceans(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	giantIceAsteroid := gamecards.Card{
		ID:   "card-giant-ice-asteroid",
		Name: "Giant Ice Asteroid",
		Type: gamecards.CardTypeEvent,
		Cost: 36,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 2, Target: "none"},
					{ResourceType: shared.ResourceTemperature, Amount: 2, Target: "none"},
					{ResourceType: shared.ResourcePlant, Amount: 6, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{giantIceAsteroid})

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
	attacker.Hand().AddCard("card-giant-ice-asteroid")

	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 10,
	})

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 36}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-giant-ice-asteroid", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Giant Ice Asteroid should play successfully")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+4, tempAfter, "Temperature should increase by 2 steps (4 degrees)")

	targetPlants := target.Resources().Get().Plants
	testutil.AssertEqual(t, 4, targetPlants, "Target should have 4 plants after removing 6")
}

// --- Trans-Neptune Probe (084) ---
// No behaviors -- just tags (science, space). Tests that a card with no behaviors can be played.

func TestTransNeptuneProbe_PlaysWithNoBehaviors(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	transNeptuneProbe := gamecards.Card{
		ID:   "card-trans-neptune-probe",
		Name: "Trans-Neptune Probe",
		Type: gamecards.CardTypeAutomated,
		Cost: 6,
		Tags: []shared.CardTag{shared.TagScience, shared.TagSpace},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{transNeptuneProbe})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-trans-neptune-probe")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-trans-neptune-probe", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Trans-Neptune Probe should play successfully with no behaviors")

	testutil.AssertFalse(t, p.Hand().HasCard("card-trans-neptune-probe"), "Card should be removed from hand")
}

// =============================================================================
// Card 106: Acquired Company
// "Increase your M€ production 3 steps."
// =============================================================================

func TestAcquiredCompany_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-acquired-company-test",
		Name: "Acquired Company",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 3, Target: "self-player"},
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
	p.Hand().AddCard("card-acquired-company-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-acquired-company-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Acquired Company should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
}

// =============================================================================
// Card 107: Media Archives
// "Gain 1 M€ for each event ever played by all players."
// =============================================================================

func TestMediaArchives_GainCreditsPerEventTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	anyPlayerTarget := "any-player"
	anywhereLocation := "anywhere"
	eventTag := shared.TagEvent

	card := gamecards.Card{
		ID:   "card-media-archives-test",
		Name: "Media Archives",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCredit,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceTag,
							Amount:       1,
							Location:     &anywhereLocation,
							Target:       &anyPlayerTarget,
							Tag:          &eventTag,
						},
					},
				},
			},
		},
	}

	eventCard1 := gamecards.Card{ID: "event1", Name: "Sabotage", Type: gamecards.CardTypeEvent, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEvent}}
	eventCard2 := gamecards.Card{ID: "event2", Name: "Virus", Type: gamecards.CardTypeEvent, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEvent}}
	eventCard3 := gamecards.Card{ID: "event3", Name: "Asteroid", Type: gamecards.CardTypeEvent, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEvent}}

	additionalCards := []gamecards.Card{card, eventCard1, eventCard2, eventCard3}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	other := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	p.PlayedCards().AddCard("event1", "Sabotage", "event", []string{"event"})
	p.PlayedCards().AddCard("event2", "Virus", "event", []string{"event"})
	other.PlayedCards().AddCard("event3", "Asteroid", "event", []string{"event"})

	p.Hand().AddCard("card-media-archives-test")

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-media-archives-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Media Archives should play successfully")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-8+3, creditsAfter,
		"Should gain 3 credits (1 per event tag, 3 events total from all players) minus 8 cost")
}

// =============================================================================
// Card 108: Open City
// "Gain 2 plants. Increase your M€ production 4 steps and decrease energy production 1 step. Place a city tile."
// Requires 12% oxygen.
// =============================================================================

func TestOpenCity_PlantsProductionAndCity(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-open-city-test",
		Name: "Open City",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 23,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(12)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
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

	testGame.GlobalParameters().SetOxygen(ctx, 12)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-open-city-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-open-city-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Open City should play successfully at 12% oxygen")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+4, prodAfter.Credits,
		"Credit production should increase by 4")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

func TestOpenCity_FailsBelowOxygenRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-open-city-fail-test",
		Name: "Open City",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 23,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(12)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 4, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-open-city-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-open-city-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Open City should fail when oxygen is below 12%")
}

// =============================================================================
// Card 109: Media Group
// "Gain 3 M€." (when played)
// =============================================================================

func TestMediaGroup_Gain3Credits(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-media-group-test",
		Name: "Media Group",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player"},
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
	p.Hand().AddCard("card-media-group-test")

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-media-group-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Media Group should play successfully")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-6+3, creditsAfter,
		"Should gain 3 credits (100 - 6 cost + 3 gain = 97)")
}

// =============================================================================
// Card 110: Business Network
// Auto: "Decrease your M€ production 1 step."
// Action: "Look at the top card and buy or discard it."
// =============================================================================

func TestBusinessNetwork_DecreaseCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-business-network-test",
		Name: "Business Network",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: -1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardBuy, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceCardPeek, Amount: 1, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 2,
	})
	p.Hand().AddCard("card-business-network-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-business-network-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Business Network should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-1, prodAfter.Credits,
		"Credit production should decrease by 1")
}

// =============================================================================
// Card 112: Bribed Committee
// "Gain 2 TR."
// =============================================================================

func TestBribedCommittee_Gain2TR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-bribed-committee-test",
		Name: "Bribed Committee",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 7,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTR, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-bribed-committee-test")

	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-bribed-committee-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Bribed Committee should play successfully")

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+2, trAfter, "TR should increase by 2")
}

// =============================================================================
// Card 113: Solar Power
// "Increase your energy production 1 step."
// =============================================================================

func TestSolarPower_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-solar-power-test",
		Name: "Solar Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-solar-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-solar-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Solar Power should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1")
}

// =============================================================================
// Card 114: Breathing Filters
// No behaviors, just VP. Requires 7% oxygen.
// =============================================================================

func TestBreathingFilters_PlaysWithOxygenRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-breathing-filters-test",
		Name: "Breathing Filters",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(7)},
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

	testGame.GlobalParameters().SetOxygen(ctx, 7)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-breathing-filters-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-breathing-filters-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Breathing Filters should play successfully at 7% oxygen")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-breathing-filters-test"),
		"Breathing Filters should be in played cards")
}

func TestBreathingFilters_FailsBelowOxygen(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-breathing-filters-fail-test",
		Name: "Breathing Filters",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(7)},
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
	p.Hand().AddCard("card-breathing-filters-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-breathing-filters-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Breathing Filters should fail with oxygen below 7%")
}

// =============================================================================
// Card 117: Geothermal Power
// "Increase your energy production 2 steps."
// =============================================================================

func TestGeothermalPower_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-geothermal-power-test",
		Name: "Geothermal Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-geothermal-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-geothermal-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Geothermal Power should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+2, prodAfter.Energy,
		"Energy production should increase by 2")
}

// =============================================================================
// Card 118: Farming
// "Gain 2 plants. Increase your M€ production 2 steps and plant production 2 steps."
// Requires 4 C or warmer.
// =============================================================================

func TestFarming_PlantsAndProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-farming-test",
		Name: "Farming",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 16,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(4)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
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

	testGame.GlobalParameters().SetTemperature(ctx, 4)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-farming-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-farming-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Farming should play successfully at 4C")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits,
		"Credit production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")
}

// =============================================================================
// Card 119: Dust Seals
// No behaviors, just VP. Requires 3 oceans or less.
// =============================================================================

func TestDustSeals_PlaysWithMaxOceanRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-dust-seals-test",
		Name: "Dust Seals",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 2,
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOceans, Max: testutil.IntPtr(3)},
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
	p.Hand().AddCard("card-dust-seals-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-dust-seals-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Dust Seals should play successfully with 0 oceans")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-dust-seals-test"),
		"Dust Seals should be in played cards")
}

// =============================================================================
// Card 122: Moss
// "Lose 1 plant. Increase plant production 1 step." Requires 3 oceans.
// =============================================================================

func TestMoss_LosePlantGainPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-moss-test",
		Name: "Moss",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOceans, Min: testutil.IntPtr(3)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
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

	testGame.GlobalParameters().SetOceans(ctx, 3)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourcePlant:  5,
	})
	p.Hand().AddCard("card-moss-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-moss-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Moss should play successfully with 3 oceans")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore-1, plantsAfter, "Should lose 1 plant")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants,
		"Plant production should increase by 1")
}

// =============================================================================
// Card 123: Industrial Center
// Action: "Spend 7 M€ to increase your steel production 1 step."
// =============================================================================

func TestIndustrialCenter_ActionSpendCreditsGainSteelProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-industrial-center"
	p.PlayedCards().AddCard(cardID, "Industrial Center", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Industrial Center",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	prodBefore := p.Resources().Production()
	creditsBefore := p.Resources().Get().Credits

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Industrial Center action should succeed")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+1, prodAfter.Steel,
		"Steel production should increase by 1")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-7, creditsAfter,
		"Should spend 7 credits")
}

// =============================================================================
// Card 125: Hackers
// "Increase your M€ production 2 steps, decrease your energy production 1 step,
//  and decrease any player's M€ production 2 steps."
// =============================================================================

func TestHackers_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-hackers-test",
		Name: "Hackers",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 3,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: -2, Target: "any-player"},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	target := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 5,
	})
	p.Hand().AddCard("card-hackers-test")

	prodBefore := p.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-hackers-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Hackers should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits,
		"Credit production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")

	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, targetProdBefore.Credits-2, targetProdAfter.Credits,
		"Target credit production should decrease by 2")
}

// =============================================================================
// Card 126: GHG Factories
// "Decrease your energy production 1 step and increase your heat production 4 steps."
// =============================================================================

func TestGHGFactories_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-ghg-factories-test",
		Name: "GHG Factories",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 4, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard("card-ghg-factories-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ghg-factories-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "GHG Factories should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
	testutil.AssertEqual(t, prodBefore.Heat+4, prodAfter.Heat,
		"Heat production should increase by 4")
}

// =============================================================================
// Card 127: Subterranean Reservoir
// "Place an ocean tile."
// =============================================================================

func TestSubterraneanReservoir_PlaceOcean(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-subterranean-reservoir-test",
		Name: "Subterranean Reservoir",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 11,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
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
	p.Hand().AddCard("card-subterranean-reservoir-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-subterranean-reservoir-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Subterranean Reservoir should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending ocean tile selection")
}

// =============================================================================
// Card 129: Zeppelins
// "Increase your M€ production 1 step for each city tile on Mars."
// Requires 5% oxygen.
// =============================================================================

func TestZeppelins_CreditProductionPerCityOnMars(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsLocation := "mars"

	card := gamecards.Card{
		ID:   "card-zeppelins-test",
		Name: "Zeppelins",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 13,
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(5)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceCityTile,
							Amount:       1,
							Location:     &marsLocation,
						},
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

	testGame.GlobalParameters().SetOxygen(ctx, 5)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-zeppelins-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-zeppelins-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Zeppelins should play successfully at 5% oxygen")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits, prodAfter.Credits,
		"Credit production should not change with 0 cities on Mars")
}

// =============================================================================
// Card 131: Decomposers
// "Add 1 microbe to this card." Requires 3% oxygen.
// =============================================================================

func TestDecomposers_AddMicrobeOnPlay(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-decomposers-test",
		Name: "Decomposers",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(3)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
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

	testGame.GlobalParameters().SetOxygen(ctx, 3)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-decomposers-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-decomposers-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Decomposers should play successfully at 3% oxygen")

	microbeStorage := p.Resources().GetCardStorage("card-decomposers-test")
	testutil.AssertEqual(t, 1, microbeStorage, "Decomposers should have 1 microbe")
}

// =============================================================================
// Card 132: Fusion Power
// "Increase your energy production 3 steps." Requires 2 power tags.
// =============================================================================

func TestFusionPower_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-fusion-power-test",
		Name: "Fusion Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 14,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower, shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(2), Tag: testutil.TagPtr(shared.TagPower)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	powerCard1 := gamecards.Card{ID: "power1", Name: "Solar Power", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 11, Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower}}
	powerCard2 := gamecards.Card{ID: "power2", Name: "Power Plant", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 4, Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower}}

	additionalCards := []gamecards.Card{card, powerCard1, powerCard2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("power1", "Solar Power", "automated", []string{"building", "power"})
	p.PlayedCards().AddCard("power2", "Power Plant", "automated", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-fusion-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-fusion-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Fusion Power should play successfully with 2 power tags")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+3, prodAfter.Energy,
		"Energy production should increase by 3")
}

func TestFusionPower_FailsWithInsufficientPowerTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-fusion-power-fail-test",
		Name: "Fusion Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 14,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower, shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(2), Tag: testutil.TagPtr(shared.TagPower)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
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
	p.Hand().AddCard("card-fusion-power-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-fusion-power-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Fusion Power should fail without 2 power tags")
}

// =============================================================================
// Card 133: Symbiotic Fungus
// Action: "Add 1 microbe to any card." Requires -14C or warmer.
// =============================================================================

func TestSymbioticFungus_ActionAddMicrobeToAnyCard(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-symbiotic-fungus"
	targetCardID := "test-target-microbe-card"
	p.PlayedCards().AddCard(cardID, "Symbiotic Fungus", "active", []string{"microbe"})
	p.PlayedCards().AddCard(targetCardID, "Decomposers", "active", []string{"microbe"})
	p.Resources().AddToStorage(targetCardID, 0)

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{
		{ID: targetCardID, Name: "Decomposers", Type: gamecards.CardTypeActive, ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe}},
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "any-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Symbiotic Fungus",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{targetCardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Symbiotic Fungus action should succeed")

	microbeStorage := p.Resources().GetCardStorage(targetCardID)
	testutil.AssertEqual(t, 1, microbeStorage, "Target card should have 1 microbe")
}

// =============================================================================
// Card 135: Advanced Ecosystems
// No behaviors, just VP. Requires 1 plant, 1 microbe, 1 animal tag.
// =============================================================================

func TestAdvancedEcosystems_PlaysWithTagRequirements(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-advanced-ecosystems-test",
		Name: "Advanced Ecosystems",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagAnimal, shared.TagMicrobe, shared.TagPlant},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(1), Tag: testutil.TagPtr(shared.TagPlant)},
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(1), Tag: testutil.TagPtr(shared.TagMicrobe)},
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(1), Tag: testutil.TagPtr(shared.TagAnimal)},
			},
		},
	}

	plantDummy := gamecards.Card{ID: "plant1", Name: "Moss", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 4, Tags: []shared.CardTag{shared.TagPlant}}
	microbeDummy := gamecards.Card{ID: "microbe1", Name: "Decomposers", Type: gamecards.CardTypeActive, Pack: "base", Cost: 5, Tags: []shared.CardTag{shared.TagMicrobe}}
	animalDummy := gamecards.Card{ID: "animal1", Name: "Fish", Type: gamecards.CardTypeActive, Pack: "base", Cost: 9, Tags: []shared.CardTag{shared.TagAnimal}}

	additionalCards := []gamecards.Card{card, plantDummy, microbeDummy, animalDummy}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("plant1", "Moss", "automated", []string{"plant"})
	p.PlayedCards().AddCard("microbe1", "Decomposers", "active", []string{"microbe"})
	p.PlayedCards().AddCard("animal1", "Fish", "active", []string{"animal"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-advanced-ecosystems-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-advanced-ecosystems-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Advanced Ecosystems should play successfully with required tags")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-advanced-ecosystems-test"),
		"Advanced Ecosystems should be in played cards")
}

// =============================================================================
// Card 136: Great Dam
// "Increase your energy production 2 steps." Requires 4 oceans.
// =============================================================================

func TestGreatDam_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-great-dam-test",
		Name: "Great Dam",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOceans, Min: testutil.IntPtr(4)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 2, Target: "self-player"},
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

	testGame.GlobalParameters().SetOceans(ctx, 4)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-great-dam-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-great-dam-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Great Dam should play successfully with 4 oceans")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+2, prodAfter.Energy,
		"Energy production should increase by 2")
}

// =============================================================================
// Card 137: Cartel
// "Increase your M€ production 1 step for each Earth tag you have, including this."
// =============================================================================

func TestCartel_CreditProductionPerEarthTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	anywhereLocation := "anywhere"
	earthTag := shared.TagEarth

	card := gamecards.Card{
		ID:   "card-cartel-test",
		Name: "Cartel",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceTag,
							Amount:       1,
							Location:     &anywhereLocation,
							Target:       &selfPlayerTarget,
							Tag:          &earthTag,
						},
					},
				},
			},
		},
	}

	earthCard1 := gamecards.Card{ID: "earth1", Name: "Media Group", Type: gamecards.CardTypeActive, Pack: "corporate-era", Cost: 6, Tags: []shared.CardTag{shared.TagEarth}}
	earthCard2 := gamecards.Card{ID: "earth2", Name: "Acquired Company", Type: gamecards.CardTypeAutomated, Pack: "corporate-era", Cost: 10, Tags: []shared.CardTag{shared.TagEarth}}

	additionalCards := []gamecards.Card{card, earthCard1, earthCard2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("earth1", "Media Group", "active", []string{"earth"})
	p.PlayedCards().AddCard("earth2", "Acquired Company", "automated", []string{"earth"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-cartel-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-cartel-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Cartel should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3 (2 existing earth tags + 1 from Cartel itself)")
}

// =============================================================================
// Card 138: Strip Mine
// "Increase your steel production 2 steps, titanium production 1 step,
//  decrease energy production 2 steps. Raise oxygen 2 steps."
// =============================================================================

func TestStripMine_ProductionAndOxygen(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-strip-mine-test",
		Name: "Strip Mine",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 25,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteelProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceOxygen, Amount: 2, Target: "none"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 3,
	})
	p.Hand().AddCard("card-strip-mine-test")

	prodBefore := p.Resources().Production()
	oxygenBefore := testGame.GlobalParameters().Oxygen()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 25}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-strip-mine-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Strip Mine should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+2, prodAfter.Steel,
		"Steel production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1")
	testutil.AssertEqual(t, prodBefore.Energy-2, prodAfter.Energy,
		"Energy production should decrease by 2")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+2, oxygenAfter,
		"Oxygen should increase by 2 steps")
}

// =============================================================================
// Card 139: Wave Power
// "Increase your energy production 1 step." Requires 3 oceans.
// =============================================================================

func TestWavePower_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-wave-power-test",
		Name: "Wave Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagPower},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOceans, Min: testutil.IntPtr(3)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
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

	testGame.GlobalParameters().SetOceans(ctx, 3)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-wave-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-wave-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Wave Power should play successfully with 3 oceans")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1")
}

// =============================================================================
// Card 140: Lava Flows
// "Raise temperature 2 steps."
// =============================================================================

func TestLavaFlows_RaiseTemperature(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-lava-flows-test",
		Name: "Lava Flows",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 18,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTemperature, Amount: 2, Target: "none"},
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
	p.Hand().AddCard("card-lava-flows-test")

	tempBefore := testGame.GlobalParameters().Temperature()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lava-flows-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lava Flows should play successfully")

	tempAfter := testGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, tempBefore+4, tempAfter,
		"Temperature should increase by 2 steps (4 degrees)")
}

// =============================================================================
// Card 141: Power Plant
// "Increase your energy production 1 step."
// =============================================================================

func TestPowerPlant_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-power-plant-test",
		Name: "Power Plant",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-power-plant-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-power-plant-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Power Plant should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1")
}

// =============================================================================
// Card 143: Large Convoy
// "Place an ocean tile. Draw 2 cards."
// =============================================================================

func TestLargeConvoy_OceanAndCardDraw(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-large-convoy-test",
		Name: "Large Convoy",
		Type: gamecards.CardTypeEvent,
		Pack: "base",
		Cost: 36,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
					{ResourceType: shared.ResourceCardDraw, Amount: 2, Target: "self-player"},
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
		shared.ResourceCredit: 200,
	})
	p.Hand().AddCard("card-large-convoy-test")

	handSizeBefore := p.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 36}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-large-convoy-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Large Convoy should play successfully")

	handSizeAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handSizeBefore-1+2, handSizeAfter,
		"Hand should have 2 more cards (minus 1 played + 2 drawn)")
}

// =============================================================================
// Card 144: Titanium Mine
// "Increase your titanium production 1 step."
// =============================================================================

func TestTitaniumMine_TitaniumProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-titanium-mine-test",
		Name: "Titanium Mine",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 7,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
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
	p.Hand().AddCard("card-titanium-mine-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-titanium-mine-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Titanium Mine should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1")
}

// =============================================================================
// Card 145: Tectonic Stress Power
// "Increase your energy production 3 steps." Requires 2 science tags.
// =============================================================================

func TestTectonicStressPower_EnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-tectonic-stress-power-test",
		Name: "Tectonic Stress Power",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 18,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(2), Tag: testutil.TagPtr(shared.TagScience)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	sciCard1 := gamecards.Card{ID: "science1", Name: "Research", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 11, Tags: []shared.CardTag{shared.TagScience}}
	sciCard2 := gamecards.Card{ID: "science2", Name: "Physics Complex", Type: gamecards.CardTypeActive, Pack: "base", Cost: 12, Tags: []shared.CardTag{shared.TagScience, shared.TagBuilding}}

	additionalCards := []gamecards.Card{card, sciCard1, sciCard2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("science1", "Research", "automated", []string{"science"})
	p.PlayedCards().AddCard("science2", "Physics Complex", "active", []string{"science", "building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-tectonic-stress-power-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-tectonic-stress-power-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Tectonic Stress Power should play successfully with 2 science tags")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+3, prodAfter.Energy,
		"Energy production should increase by 3")
}

// =============================================================================
// Card 146: Nitrophilic Moss
// "Lose 2 plants. Increase your plant production 2 steps." Requires 3 oceans.
// =============================================================================

func TestNitrophilicMoss_LosePlantsGainPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-nitrophilic-moss-test",
		Name: "Nitrophilic Moss",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagPlant},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOceans, Min: testutil.IntPtr(3)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
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

	testGame.GlobalParameters().SetOceans(ctx, 3)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourcePlant:  5,
	})
	p.Hand().AddCard("card-nitrophilic-moss-test")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-nitrophilic-moss-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nitrophilic Moss should play successfully with 3 oceans")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore-2, plantsAfter, "Should lose 2 plants")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")
}

// =============================================================================
// Card 147: Herbivores
// "Add 1 animal to this card. Decrease any player's plant production 1 step."
// Requires 8% oxygen.
// =============================================================================

func TestHerbivores_AddAnimalAndDecreaseTargetPlantProd(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-herbivores-test",
		Name: "Herbivores",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagAnimal},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(8)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	additionalCards := []gamecards.Card{card}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	target := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetOxygen(ctx, 8)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 3,
	})
	p.Hand().AddCard("card-herbivores-test")

	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-herbivores-test", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Herbivores should play successfully at 8% oxygen")

	animalStorage := p.Resources().GetCardStorage("card-herbivores-test")
	testutil.AssertEqual(t, 1, animalStorage, "Herbivores should have 1 animal")

	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, targetProdBefore.Plants-1, targetProdAfter.Plants,
		"Target plant production should decrease by 1")
}

// =============================================================================
// Card 148: Insects
// "Increase your plant production 1 step for each plant tag you have."
// Requires 6% oxygen.
// =============================================================================

func TestInsects_PlantProductionPerPlantTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	anywhereLocation := "anywhere"
	plantTag := shared.TagPlant

	card := gamecards.Card{
		ID:   "card-insects-test",
		Name: "Insects",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementOxygen, Min: testutil.IntPtr(6)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourcePlantProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceTag,
							Amount:       1,
							Location:     &anywhereLocation,
							Target:       &selfPlayerTarget,
							Tag:          &plantTag,
						},
					},
				},
			},
		},
	}

	plantDummy1 := gamecards.Card{ID: "plant1", Name: "Moss", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 4, Tags: []shared.CardTag{shared.TagPlant}}
	plantDummy2 := gamecards.Card{ID: "plant2", Name: "Farming", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 16, Tags: []shared.CardTag{shared.TagPlant}}
	plantDummy3 := gamecards.Card{ID: "plant3", Name: "Nitrophilic Moss", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 8, Tags: []shared.CardTag{shared.TagPlant}}

	additionalCards := []gamecards.Card{card, plantDummy1, plantDummy2, plantDummy3}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	testGame.GlobalParameters().SetOxygen(ctx, 6)

	p.PlayedCards().AddCard("plant1", "Moss", "automated", []string{"plant"})
	p.PlayedCards().AddCard("plant2", "Farming", "automated", []string{"plant"})
	p.PlayedCards().AddCard("plant3", "Nitrophilic Moss", "automated", []string{"plant"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-insects-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-insects-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Insects should play successfully at 6% oxygen")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+3, prodAfter.Plants,
		"Plant production should increase by 3 (1 per plant tag, 3 plant tags in play)")
}

// =============================================================================
// Card 151: Investment Loan
// "Gain 10 M€. Decrease your M€ production 1 step."
// =============================================================================

func TestInvestmentLoan_GainCreditsLoseProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-investment-loan-test",
		Name: "Investment Loan",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 3,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 10, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: -1, Target: "self-player"},
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
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 2,
	})
	p.Hand().AddCard("card-investment-loan-test")

	creditsBefore := p.Resources().Get().Credits
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-investment-loan-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Investment Loan should play successfully")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-3+10, creditsAfter,
		"Should gain 10 credits minus 3 cost = net +7")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits-1, prodAfter.Credits,
		"Credit production should decrease by 1")
}

// =============================================================================
// Card 154: Caretaker Contract
// Action: "Spend 8 heat to gain 1 TR." Requires 0C or warmer.
// =============================================================================

func TestCaretakerContract_ActionSpendHeatGainTR(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	testGame.GlobalParameters().SetTemperature(ctx, 0)

	cardID := "test-caretaker-contract"
	p.PlayedCards().AddCard(cardID, "Caretaker Contract", "active", []string{})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceHeat: 15,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceHeat, Amount: 8, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTR, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Caretaker Contract",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	trBefore := p.Resources().TerraformRating()
	heatBefore := p.Resources().Get().Heat

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Caretaker Contract action should succeed")

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1")

	heatAfter := p.Resources().Get().Heat
	testutil.AssertEqual(t, heatBefore-8, heatAfter, "Should spend 8 heat")
}

// =============================================================================
// Card 155: Designed Microorganisms
// "Increase your plant production 2 steps." Requires -14C or colder.
// =============================================================================

func TestDesignedMicroorganisms_PlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-designed-microorganisms-test",
		Name: "Designed Microorganisms",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 16,
		Tags: []shared.CardTag{shared.TagMicrobe, shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTemperature, Max: testutil.IntPtr(-14)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-designed-microorganisms-test")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-designed-microorganisms-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Designed Microorganisms should play successfully at -30C (below -14C)")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants,
		"Plant production should increase by 2")
}

func TestDesignedMicroorganisms_FailsAboveMaxTemp(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-designed-microorganisms-fail-test",
		Name: "Designed Microorganisms",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 16,
		Tags: []shared.CardTag{shared.TagMicrobe, shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTemperature, Max: testutil.IntPtr(-14)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
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

	testGame.GlobalParameters().SetTemperature(ctx, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-designed-microorganisms-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-designed-microorganisms-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Designed Microorganisms should fail when temperature is above -14C")
}

// =============================================================================
// Card 156: Standard Technology
// "Gain 3 M€." (when played)
// =============================================================================

func TestStandardTechnology_Gain3Credits(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-standard-technology-test",
		Name: "Standard Technology",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player"},
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
	p.Hand().AddCard("card-standard-technology-test")

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-standard-technology-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Standard Technology should play successfully")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-6+3, creditsAfter,
		"Should gain 3 credits (100 - 6 cost + 3 gain = 97)")
}

// =============================================================================
// Card 153: Adaptation Technology
// "Gain +2 global parameter requirement lenience."
// =============================================================================

func TestAdaptationTechnology_GlobalParameterLenience(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-adaptation-technology-test",
		Name: "Adaptation Technology",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceGlobalParameterLenience, Amount: 2, Target: "self-player"},
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
	p.Hand().AddCard("card-adaptation-technology-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-adaptation-technology-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Adaptation Technology should play successfully")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-adaptation-technology-test"),
		"Adaptation Technology should be in played cards")
}

// =============================================================================
// Card 150: Anti-Gravity Technology
// "All cards cost 2 M€ less." Requires 7 science tags.
// =============================================================================

func TestAntiGravityTechnology_Discount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-anti-gravity-technology-test",
		Name: "Anti-Gravity Technology",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 14,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(7), Tag: testutil.TagPtr(shared.TagScience)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	scienceCards := make([]gamecards.Card, 7)
	for i := 0; i < 7; i++ {
		scienceCards[i] = gamecards.Card{
			ID:   "science-dummy-" + string(rune('a'+i)),
			Name: "Science Dummy",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 1,
			Tags: []shared.CardTag{shared.TagScience},
		}
	}
	additionalCards := append([]gamecards.Card{card}, scienceCards...)
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	for i := 0; i < 7; i++ {
		cardIDStr := "science-dummy-" + string(rune('a'+i))
		p.PlayedCards().AddCard(cardIDStr, "Science Dummy", "automated", []string{"science"})
	}

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-anti-gravity-technology-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-anti-gravity-technology-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Anti-Gravity Technology should play successfully with 7 science tags")

	testutil.AssertTrue(t, p.PlayedCards().Contains("card-anti-gravity-technology-test"),
		"Anti-Gravity Technology should be in played cards")
}

func TestAntiGravityTechnology_FailsWithInsufficientScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	card := gamecards.Card{
		ID:   "card-anti-gravity-fail-test",
		Name: "Anti-Gravity Technology",
		Type: gamecards.CardTypeActive,
		Pack: "base",
		Cost: 14,
		Tags: []shared.CardTag{shared.TagScience},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementTags, Min: testutil.IntPtr(7), Tag: testutil.TagPtr(shared.TagScience)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player"},
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

	p.PlayedCards().AddCard("science-1", "Research", "automated", []string{"science"})
	p.PlayedCards().AddCard("science-2", "Physics Complex", "active", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-anti-gravity-fail-test")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-anti-gravity-fail-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Anti-Gravity Technology should fail without 7 science tags")
}

// =============================================================================
// Card 158: Industrial Microbes
// "Increase your energy production and your steel production 1 step each."
// =============================================================================

func TestIndustrialMicrobes_IncreasesEnergyAndSteelProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	industrialMicrobes := gamecards.Card{
		ID:   "card-industrial-microbes",
		Name: "Industrial Microbes",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 12,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{industrialMicrobes})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-industrial-microbes")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-industrial-microbes", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Industrial Microbes should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+1, prodAfter.Steel, "Steel production should increase by 1")
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy, "Energy production should increase by 1")
}

// =============================================================================
// Card 159: Lichen
// "Requires -24C or warmer. Increase your plant production 1 step."
// =============================================================================

func TestLichen_IncreasesPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	lichen := gamecards.Card{
		ID:   "card-lichen",
		Name: "Lichen",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 7,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{lichen})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Set temperature to -24 (3 steps from -30)
	testGame.GlobalParameters().IncreaseTemperature(ctx, 3)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-lichen")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lichen", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lichen should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants, "Plant production should increase by 1")
}

// =============================================================================
// Card 162: Imported GHG
// "Increase your heat production 1 step and gain 3 heat."
// =============================================================================

func TestImportedGHG_IncreasesHeatProductionAndGainsHeat(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	importedGHG := gamecards.Card{
		ID:   "card-imported-ghg",
		Name: "Imported GHG",
		Type: gamecards.CardTypeEvent,
		Pack: "base-game",
		Cost: 7,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeat, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{importedGHG})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-imported-ghg")

	heatBefore := p.Resources().Get().Heat
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-imported-ghg", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Imported GHG should play successfully")

	heatAfter := p.Resources().Get().Heat
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, heatBefore+3, heatAfter, "Should gain 3 heat")
	testutil.AssertEqual(t, prodBefore.Heat+1, prodAfter.Heat, "Heat production should increase by 1")
}

// =============================================================================
// Card 164: Micro-Mills
// "Increase your heat production 1 step."
// =============================================================================

func TestMicroMills_IncreasesHeatProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	microMills := gamecards.Card{
		ID:   "card-micro-mills",
		Name: "Micro-Mills",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 3,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{microMills})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-micro-mills")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-micro-mills", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Micro-Mills should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+1, prodAfter.Heat, "Heat production should increase by 1")
}

// =============================================================================
// Card 165: Magnetic Field Generators
// "Decrease your energy production 4 steps and increase your plant production 2 steps. Raise your TR 3 steps."
// =============================================================================

func TestMagneticFieldGenerators_ProductionAndTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	magneticFieldGenerators := gamecards.Card{
		ID:   "card-magnetic-field-generators",
		Name: "Magnetic Field Generators",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 20,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -4, Target: "self-player"},
					{ResourceType: shared.ResourceTR, Amount: 3, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{magneticFieldGenerators})

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
		shared.ResourceEnergyProduction: 4,
	})
	p.Hand().AddCard("card-magnetic-field-generators")

	prodBefore := p.Resources().Production()
	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 20}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-magnetic-field-generators", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Magnetic Field Generators should play successfully")

	prodAfter := p.Resources().Production()
	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, prodBefore.Energy-4, prodAfter.Energy, "Energy production should decrease by 4")
	testutil.AssertEqual(t, prodBefore.Plants+2, prodAfter.Plants, "Plant production should increase by 2")
	testutil.AssertEqual(t, trBefore+3, trAfter, "TR should increase by 3")
}

// =============================================================================
// Card 167: Import Of Advanced GHG
// "Increase your heat production 2 steps."
// =============================================================================

func TestImportOfAdvancedGHG_IncreasesHeatProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	importAdvGHG := gamecards.Card{
		ID:   "card-import-advanced-ghg",
		Name: "Import Of Advanced GHG",
		Type: gamecards.CardTypeEvent,
		Pack: "base-game",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{importAdvGHG})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-import-advanced-ghg")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-import-advanced-ghg", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Import Of Advanced GHG should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+2, prodAfter.Heat, "Heat production should increase by 2")
}

// =============================================================================
// Card 169: Tundra Farming
// "Requires -6C or warmer. Increase your plant production 1 step and your M$ production 2 steps. Gain 1 plant."
// =============================================================================

func TestTundraFarming_ProductionAndPlant(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	tundraFarming := gamecards.Card{
		ID:   "card-tundra-farming",
		Name: "Tundra Farming",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 16,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{tundraFarming})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Temperature needs to be -6 or warmer. Default is -30. Need 12 steps (each step is +2).
	testGame.GlobalParameters().IncreaseTemperature(ctx, 12)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-tundra-farming")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-tundra-farming", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Tundra Farming should play successfully")

	plantsAfter := p.Resources().Get().Plants
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, plantsBefore+1, plantsAfter, "Should gain 1 plant")
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits, "Credit production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants, "Plant production should increase by 1")
}

// =============================================================================
// Card 170: Aerobraked Ammonia Asteroid
// "Add 2 microbes to another card. Increase your heat production 3 steps and your plant production 1 step."
// =============================================================================

func TestAerobrakedAmmoniaAsteroid_ProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	aerobrakedAsteroid := gamecards.Card{
		ID:   "card-aerobraked-ammonia-asteroid",
		Name: "Aerobraked Ammonia Asteroid",
		Type: gamecards.CardTypeEvent,
		Pack: "base-game",
		Cost: 26,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: 3, Target: "self-player"},
					{ResourceType: shared.ResourceMicrobe, Amount: 2, Target: "any-card"},
				},
			},
		},
	}

	microbeHost := gamecards.Card{
		ID: "card-microbe-host-b4", Name: "Microbe Host B4", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagMicrobe},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{aerobrakedAsteroid, microbeHost})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.PlayedCards().AddCard("card-microbe-host-b4", "Microbe Host B4", "active", []string{"microbe"})
	p.Resources().AddToStorage("card-microbe-host-b4", 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-aerobraked-ammonia-asteroid")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 26}
	targetCardID := "card-microbe-host-b4"
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-aerobraked-ammonia-asteroid", payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Aerobraked Ammonia Asteroid should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+3, prodAfter.Heat, "Heat production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants, "Plant production should increase by 1")

	microbeStorage := p.Resources().GetCardStorage("card-microbe-host-b4")
	testutil.AssertEqual(t, 2, microbeStorage, "Should add 2 microbes to host card")
}

// =============================================================================
// Card 171: Magnetic Field Dome
// "Decrease your energy production 2 steps and increase your plant production 1 step. Raise your terraform rating 1 step."
// =============================================================================

func TestMagneticFieldDome_ProductionAndTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	magneticFieldDome := gamecards.Card{
		ID:   "card-magnetic-field-dome",
		Name: "Magnetic Field Dome",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -2, Target: "self-player"},
					{ResourceType: shared.ResourceTR, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{magneticFieldDome})

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
	p.Hand().AddCard("card-magnetic-field-dome")

	prodBefore := p.Resources().Production()
	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-magnetic-field-dome", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Magnetic Field Dome should play successfully")

	prodAfter := p.Resources().Production()
	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, prodBefore.Energy-2, prodAfter.Energy, "Energy production should decrease by 2")
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants, "Plant production should increase by 1")
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1")
}

// =============================================================================
// Card 175: Satellites
// "Increase your M$ production 1 step for each space tag you have, including this."
// =============================================================================

func TestSatellites_CreditProductionPerSpaceTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	satellites := gamecards.Card{
		ID:   "card-satellites",
		Name: "Satellites",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagSpace),
							Amount:       1,
							Target:       testutil.StrPtr("self-player"),
							Tag:          testutil.TagPtr(shared.TagSpace),
						},
					},
				},
			},
		},
	}

	spaceCard1 := gamecards.Card{ID: "card-space-1-b4", Name: "Space Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagSpace}, Cost: 0}
	spaceCard2 := gamecards.Card{ID: "card-space-2-b4", Name: "Space Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagSpace}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{satellites, spaceCard1, spaceCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Play 2 space-tagged cards first
	p.PlayedCards().AddCard("card-space-1-b4", "Space Card 1", "automated", []string{"space"})
	p.PlayedCards().AddCard("card-space-2-b4", "Space Card 2", "automated", []string{"space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-satellites")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-satellites", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Satellites should play successfully")

	prodAfter := p.Resources().Production()
	// 2 existing space tags + 1 from Satellites itself = 3 credit production
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3 (1 per each of 2 existing + 1 self space tag)")
}

func TestSatellites_ZeroExistingSpaceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	satellites := gamecards.Card{
		ID:   "card-satellites",
		Name: "Satellites",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagSpace),
							Amount:       1,
							Target:       testutil.StrPtr("self-player"),
							Tag:          testutil.TagPtr(shared.TagSpace),
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{satellites})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-satellites")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-satellites", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Satellites with no prior space tags should play successfully")

	prodAfter := p.Resources().Production()
	// Only 1 space tag from Satellites itself = 1 credit production
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits,
		"Credit production should increase by 1 (only self space tag)")
}

// =============================================================================
// Card 176: Noctis Farming
// "Requires -20C or warmer. Increase your M$ production 1 step and gain 2 plants."
// =============================================================================

func TestNoctisFarming_ProductionAndPlants(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	noctisFarming := gamecards.Card{
		ID:   "card-noctis-farming",
		Name: "Noctis Farming",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlant, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{noctisFarming})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Temperature needs to be -20 or warmer. Default is -30. Need 5 steps (each step is +2).
	testGame.GlobalParameters().IncreaseTemperature(ctx, 5)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-noctis-farming")

	plantsBefore := p.Resources().Get().Plants
	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-noctis-farming", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Noctis Farming should play successfully")

	plantsAfter := p.Resources().Get().Plants
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, plantsBefore+2, plantsAfter, "Should gain 2 plants")
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits, "Credit production should increase by 1")
}

// =============================================================================
// Card 179: Soil Factory
// "Decrease your energy production 1 step and increase your plant production 1 step."
// =============================================================================

func TestSoilFactory_SwapsEnergyForPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	soilFactory := gamecards.Card{
		ID:   "card-soil-factory",
		Name: "Soil Factory",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{soilFactory})

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
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-soil-factory")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-soil-factory", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Soil Factory should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy, "Energy production should decrease by 1")
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants, "Plant production should increase by 1")
}

// =============================================================================
// Card 180: Fuel Factory
// "Decrease your energy production 1 step and increase your titanium and your M$ production 1 step each."
// =============================================================================

func TestFuelFactory_SwapsEnergyForTitaniumAndCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	fuelFactory := gamecards.Card{
		ID:   "card-fuel-factory",
		Name: "Fuel Factory",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{fuelFactory})

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
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-fuel-factory")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-fuel-factory", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Fuel Factory should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy, "Energy production should decrease by 1")
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium, "Titanium production should increase by 1")
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits, "Credit production should increase by 1")
}

// =============================================================================
// Card 186: Rad-Suits
// "Requires 2 cities in play. Increase your M$ production 1 step."
// =============================================================================

func TestRadSuits_IncreasesCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	radSuits := gamecards.Card{
		ID:   "card-rad-suits",
		Name: "Rad-Suits",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 6,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{radSuits})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-rad-suits")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-rad-suits", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Rad-Suits should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits, "Credit production should increase by 1")
}

// =============================================================================
// Card 196: Lagrange Observatory
// "Draw 1 card."
// =============================================================================

func TestLagrangeObservatory_DrawsCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	lagrange := gamecards.Card{
		ID:   "card-lagrange-observatory",
		Name: "Lagrange Observatory",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 9,
		Tags: []shared.CardTag{shared.TagScience, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{lagrange})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-lagrange-observatory")

	handBefore := p.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-lagrange-observatory", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Lagrange Observatory should play successfully")

	handAfter := p.Hand().CardCount()
	// Played 1 card (-1) and drew 1 card (+1) = net 0
	testutil.AssertEqual(t, handBefore, handAfter,
		"Hand size should remain same (played 1, drew 1)")
}

// =============================================================================
// Card 198: Immigration Shuttles
// "Increase your M$ production 5 steps."
// =============================================================================

func TestImmigrationShuttles_IncreasesCreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	immigrationShuttles := gamecards.Card{
		ID:   "card-immigration-shuttles",
		Name: "Immigration Shuttles",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 31,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 5, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{immigrationShuttles})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-immigration-shuttles")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 31}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-immigration-shuttles", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Immigration Shuttles should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+5, prodAfter.Credits, "Credit production should increase by 5")
}

// =============================================================================
// Card 203: Soletta
// "Increase your heat production 7 steps."
// =============================================================================

func TestSoletta_IncreasesHeatProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	soletta := gamecards.Card{
		ID:   "card-soletta",
		Name: "Soletta",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 35,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: 7, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{soletta})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-soletta")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 35}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-soletta", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Soletta should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+7, prodAfter.Heat, "Heat production should increase by 7")
}

// =============================================================================
// Card 204: Technology Demonstration
// "Draw 2 cards."
// =============================================================================

func TestTechnologyDemonstration_DrawsTwoCards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	techDemo := gamecards.Card{
		ID:   "card-technology-demonstration",
		Name: "Technology Demonstration",
		Type: gamecards.CardTypeEvent,
		Pack: "corporate-era",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagScience, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{techDemo})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-technology-demonstration")

	handBefore := p.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-technology-demonstration", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Technology Demonstration should play successfully")

	handAfter := p.Hand().CardCount()
	// Played 1 card (-1) and drew 2 cards (+2) = net +1
	testutil.AssertEqual(t, handBefore+1, handAfter,
		"Hand size should increase by 1 (played 1, drew 2)")
}

// =============================================================================
// Card 205: Rad-Chem Factory
// "Decrease your energy production 1 step. Raise your terraform rating 2 steps."
// =============================================================================

func TestRadChemFactory_ProductionAndTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	radChemFactory := gamecards.Card{
		ID:   "card-rad-chem-factory",
		Name: "Rad-Chem Factory",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{ResourceType: shared.ResourceTR, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{radChemFactory})

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
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-rad-chem-factory")

	prodBefore := p.Resources().Production()
	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-rad-chem-factory", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Rad-Chem Factory should play successfully")

	prodAfter := p.Resources().Production()
	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy, "Energy production should decrease by 1")
	testutil.AssertEqual(t, trBefore+2, trAfter, "TR should increase by 2")
}

// =============================================================================
// Card 207: Medical Lab
// "Increase your M$ production 1 step for every 2 building tags you have, including this."
// =============================================================================

func TestMedicalLab_CreditProductionPerTwoBuildingTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	medicalLab := gamecards.Card{
		ID:   "card-medical-lab",
		Name: "Medical Lab",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagBuilding),
							Amount:       2,
							Target:       testutil.StrPtr("self-player"),
							Tag:          testutil.TagPtr(shared.TagBuilding),
						},
					},
				},
			},
		},
	}

	buildingCard1 := gamecards.Card{ID: "card-building-1-b4", Name: "Building Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}
	buildingCard2 := gamecards.Card{ID: "card-building-2-b4", Name: "Building Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}
	buildingCard3 := gamecards.Card{ID: "card-building-3-b4", Name: "Building Card 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{medicalLab, buildingCard1, buildingCard2, buildingCard3})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Play 3 building-tagged cards first
	p.PlayedCards().AddCard("card-building-1-b4", "Building Card 1", "automated", []string{"building"})
	p.PlayedCards().AddCard("card-building-2-b4", "Building Card 2", "automated", []string{"building"})
	p.PlayedCards().AddCard("card-building-3-b4", "Building Card 3", "automated", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-medical-lab")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-medical-lab", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Medical Lab should play successfully")

	prodAfter := p.Resources().Production()
	// 3 existing building tags + 1 from Medical Lab itself = 4 building tags. 4 / 2 = 2 credit production
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits,
		"Credit production should increase by 2 (4 building tags / 2)")
}

func TestMedicalLab_OddNumberOfBuildingTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	medicalLab := gamecards.Card{
		ID:   "card-medical-lab",
		Name: "Medical Lab",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceCreditProduction,
						Amount:       1,
						Target:       "self-player",
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagBuilding),
							Amount:       2,
							Target:       testutil.StrPtr("self-player"),
							Tag:          testutil.TagPtr(shared.TagBuilding),
						},
					},
				},
			},
		},
	}

	buildingCard1 := gamecards.Card{ID: "card-building-odd-1", Name: "Building Card Odd 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}
	buildingCard2 := gamecards.Card{ID: "card-building-odd-2", Name: "Building Card Odd 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{medicalLab, buildingCard1, buildingCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Play 2 building-tagged cards first
	p.PlayedCards().AddCard("card-building-odd-1", "Building Card Odd 1", "automated", []string{"building"})
	p.PlayedCards().AddCard("card-building-odd-2", "Building Card Odd 2", "automated", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-medical-lab")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-medical-lab", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Medical Lab with odd building tags should play successfully")

	prodAfter := p.Resources().Production()
	// 2 existing + 1 from self = 3 building tags. 3 / 2 = 1 (floor division)
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits,
		"Credit production should increase by 1 (3 building tags / 2 = 1, floor division)")
}

// =============================================================================
// Card 157: Nitrite Reducing Bacteria (active card with choices)
// "Action: Add 1 microbe to this card, or remove 3 microbes to increase your TR 1 step. Add 3 microbes to this card."
// =============================================================================

func TestNitriteReducingBacteria_OnPlay_AddsThreeMicrobes(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	nitriteReducing := gamecards.Card{
		ID:   "card-nitrite-reducing-bacteria",
		Name: "Nitrite Reducing Bacteria",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 3, Target: "self-card"},
				},
			},
			nitriteReducingBacteriaBehavior(),
		},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{nitriteReducing})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-nitrite-reducing-bacteria")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-nitrite-reducing-bacteria", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nitrite Reducing Bacteria should play successfully")

	microbeStorage := p.Resources().GetCardStorage("card-nitrite-reducing-bacteria")
	testutil.AssertEqual(t, 3, microbeStorage, "Should have 3 microbes on card after playing")
}

func TestNitriteReducingBacteria_Action_AddMicrobe(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-nitrite-reducing-bacteria"
	nitriteReducing := gamecards.Card{
		ID:   cardID,
		Name: "Nitrite Reducing Bacteria",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 3, Target: "self-card"},
				},
			},
			nitriteReducingBacteriaBehavior(),
		},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{nitriteReducing})

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

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nitrite Reducing Bacteria should play successfully")

	testutil.AssertEqual(t, 3, p.Resources().GetCardStorage(cardID), "Should have 3 microbes after playing")

	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 1, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add microbe) should succeed")

	testutil.AssertEqual(t, 4, p.Resources().GetCardStorage(cardID), "Card should have 4 microbes after adding 1")
}

func TestNitriteReducingBacteria_Action_Remove3MicrobesForTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-nitrite-reducing-bacteria"
	nitriteReducing := gamecards.Card{
		ID:   cardID,
		Name: "Nitrite Reducing Bacteria",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 11,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 3, Target: "self-card"},
				},
			},
			nitriteReducingBacteriaBehavior(),
		},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{nitriteReducing})

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

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nitrite Reducing Bacteria should play successfully")

	// Card starts with 3 microbes from auto-trigger, add 2 more for the test
	p.Resources().AddToStorage(cardID, 2)
	testutil.AssertEqual(t, 5, p.Resources().GetCardStorage(cardID), "Should have 5 microbes before action")

	trBefore := p.Resources().TerraformRating()

	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 1, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 1 (remove 3 microbes for TR) should succeed")

	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Card should have 2 microbes after removing 3 from 5")
	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1")
}

func TestNitriteReducingBacteria_Action_FailsWithInsufficientMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-nitrite-reducing-bacteria"
	p.PlayedCards().AddCard(cardID, "Nitrite Reducing Bacteria", "active", []string{"microbe"})
	p.Resources().AddToStorage(cardID, 2) // Only 2 microbes, need 3

	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Nitrite Reducing Bacteria",
			BehaviorIndex: 1,
			Behavior:      nitriteReducingBacteriaBehavior(),
		},
	})

	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Choice 1 should fail with only 2 microbes (need 3)")
}

// =============================================================================
// Card 177: Water Splitting Plant (active card action)
// "Action: Spend 3 energy to raise oxygen 1 step. Requires 2 ocean tiles."
// =============================================================================

func TestWaterSplittingPlant_SpendEnergyToRaiseOxygen(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-water-splitting-plant"
	p.PlayedCards().AddCard(cardID, "Water Splitting Plant", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 5,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 3, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Splitting Plant",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	oxygenBefore := testGame.GlobalParameters().Oxygen()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Water Splitting Plant action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Energy, "Should have 2 energy after spending 3")

	oxygenAfter := testGame.GlobalParameters().Oxygen()
	testutil.AssertEqual(t, oxygenBefore+1, oxygenAfter, "Oxygen should increase by 1")
}

func TestWaterSplittingPlant_FailsWithoutEnoughEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-water-splitting-plant"
	p.PlayedCards().AddCard(cardID, "Water Splitting Plant", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 2, // Only 2, need 3
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 3, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOxygen, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Splitting Plant",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Water Splitting Plant should fail without enough energy")
}

// =============================================================================
// Card 184: Livestock (active card action)
// "Action: Add 1 animal to this card. Requires 9% oxygen."
// "Decrease your plant production 1 step and increase your M$ production 2 steps."
// =============================================================================

func TestLivestock_OnPlay_ChangesProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	livestock := gamecards.Card{
		ID:   "card-livestock",
		Name: "Livestock",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagAnimal},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
				},
			},
		},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{livestock})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Oxygen needs to be at 9 for the requirement
	testGame.GlobalParameters().IncreaseOxygen(ctx, 9)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 1,
	})
	p.Hand().AddCard("card-livestock")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-livestock", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Livestock should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits, "Credit production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Plants-1, prodAfter.Plants, "Plant production should decrease by 1")
}

func TestLivestock_Action_AddAnimal(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-livestock"
	livestock := gamecards.Card{
		ID:   cardID,
		Name: "Livestock",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 13,
		Tags: []shared.CardTag{shared.TagAnimal},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "self-player"},
				},
			},
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
				},
			},
		},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{livestock})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Oxygen needs to be at 9 for the requirement (same as OnPlay test)
	testGame.GlobalParameters().IncreaseOxygen(ctx, 9)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 1,
	})
	p.Hand().AddCard(cardID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Livestock should play successfully")

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 1, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Livestock action should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 animal after action")
}

// =============================================================================
// Card 202: Underground Detonations (active card action)
// "Action: Spend 10 M$ to increase your heat production 2 steps."
// =============================================================================

func TestUndergroundDetonations_SpendCreditsForHeatProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-underground-detonations"
	p.PlayedCards().AddCard(cardID, "Underground Detonations", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 10, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Underground Detonations",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	prodBefore := p.Resources().Production()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Underground Detonations action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 10, resources.Credits, "Should have 10 credits after spending 10")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+2, prodAfter.Heat, "Heat production should increase by 2")
}

func TestUndergroundDetonations_FailsWithInsufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-underground-detonations"
	p.PlayedCards().AddCard(cardID, "Underground Detonations", "active", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 5, // Only 5, need 10
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 10, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceHeatProduction, Amount: 2, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Underground Detonations",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Underground Detonations should fail without enough credits")
}

// =============================================================================
// Card 208: AI Central (active card action)
// "Action: Draw 2 cards. Requires 3 science tags to play. Decrease your energy production 1 step."
// =============================================================================

func TestAICentral_OnPlay_DecreasesEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	aiCentral := gamecards.Card{
		ID:   "card-ai-central",
		Name: "AI Central",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 21,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
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
					{ResourceType: shared.ResourceCardDraw, Amount: 2, Target: "self-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{aiCentral})

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
	p.Hand().AddCard("card-ai-central")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ai-central", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "AI Central should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy, "Energy production should decrease by 1")
}

func TestAICentral_Action_DrawTwoCards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := "card-ai-central"
	aiCentral := gamecards.Card{
		ID:   cardID,
		Name: "AI Central",
		Type: gamecards.CardTypeActive,
		Pack: "corporate-era",
		Cost: 21,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
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
					{ResourceType: shared.ResourceCardDraw, Amount: 2, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{aiCentral})

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
	p.Hand().AddCard(cardID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "AI Central should play successfully")

	handBefore := p.Hand().CardCount()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), cardID, 1, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "AI Central action should succeed")

	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore+2, handAfter, "Hand should increase by 2 from drawing 2 cards")
}

// =============================================================================
// Card 160: Power Supply Consortium
// "Requires 2 power tags. Decrease any energy production 1 step and increase your own 1 step."
// =============================================================================

func TestPowerSupplyConsortium_StealEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	powerSupplyConsortium := gamecards.Card{
		ID:   "card-power-supply-consortium",
		Name: "Power Supply Consortium",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{powerSupplyConsortium})

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
	attacker.Hand().AddCard("card-power-supply-consortium")

	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 3,
	})

	attackerProdBefore := attacker.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-power-supply-consortium", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Power Supply Consortium should play successfully")

	attackerProdAfter := attacker.Resources().Production()
	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, attackerProdBefore.Energy+1, attackerProdAfter.Energy,
		"Attacker energy production should increase by 1")
	testutil.AssertEqual(t, targetProdBefore.Energy-1, targetProdAfter.Energy,
		"Target energy production should decrease by 1")
}

// =============================================================================
// Card 201: Energy Tapping
// "Decrease any energy production 1 step and increase your own 1 step."
// =============================================================================

func TestEnergyTapping_StealEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	energyTapping := gamecards.Card{
		ID:   "card-energy-tapping",
		Name: "Energy Tapping",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 3,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{energyTapping})

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
	attacker.Hand().AddCard("card-energy-tapping")

	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})

	attackerProdBefore := attacker.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-energy-tapping", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Energy Tapping should play successfully")

	attackerProdAfter := attacker.Resources().Production()
	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, attackerProdBefore.Energy+1, attackerProdAfter.Energy,
		"Attacker energy production should increase by 1")
	testutil.AssertEqual(t, targetProdBefore.Energy-1, targetProdAfter.Energy,
		"Target energy production should decrease by 1")
}

func TestEnergyTapping_SoloMode(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	energyTapping := gamecards.Card{
		ID:   "card-energy-tapping",
		Name: "Energy Tapping",
		Type: gamecards.CardTypeAutomated,
		Pack: "corporate-era",
		Cost: 3,
		Tags: []shared.CardTag{shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{energyTapping})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-energy-tapping")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-energy-tapping", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Energy Tapping should play successfully in solo mode")

	prodAfter := p.Resources().Production()
	// Self-player (+1) applied, any-player (-1) skipped in solo
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy,
		"Energy production should increase by 1 in solo (any-player decrease skipped)")
}

// =============================================================================
// Card 178: Heat Trappers
// "Decrease any heat production 2 steps and increase your energy production 1 step."
// =============================================================================

func TestHeatTrappers_StealHeatProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	heatTrappers := gamecards.Card{
		ID:   "card-heat-trappers",
		Name: "Heat Trappers",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 6,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
					{ResourceType: shared.ResourceHeatProduction, Amount: -2, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{heatTrappers})

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
	attacker.Hand().AddCard("card-heat-trappers")

	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceHeatProduction: 5,
	})

	attackerProdBefore := attacker.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-heat-trappers", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Heat Trappers should play successfully")

	attackerProdAfter := attacker.Resources().Production()
	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, attackerProdBefore.Energy+1, attackerProdAfter.Energy,
		"Attacker energy production should increase by 1")
	testutil.AssertEqual(t, targetProdBefore.Heat-2, targetProdAfter.Heat,
		"Target heat production should decrease by 2")
}

// =============================================================================
// Card 183: Biomass Combustors
// "Requires 6% oxygen. Decrease any plant production 1 step and increase your energy production 2 steps."
// =============================================================================

func TestBiomassCombustors_StealPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	biomassCombustors := gamecards.Card{
		ID:   "card-biomass-combustors",
		Name: "Biomass Combustors",
		Type: gamecards.CardTypeAutomated,
		Pack: "base-game",
		Cost: 4,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagPower},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourcePlantProduction, Amount: -1, Target: "any-player"},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{biomassCombustors})

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	// Oxygen needs to be 6
	testGame.GlobalParameters().IncreaseOxygen(ctx, 6)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-biomass-combustors")

	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 3,
	})

	attackerProdBefore := attacker.Resources().Production()
	targetProdBefore := target.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-biomass-combustors", payment, nil, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Biomass Combustors should play successfully")

	attackerProdAfter := attacker.Resources().Production()
	targetProdAfter := target.Resources().Production()
	testutil.AssertEqual(t, attackerProdBefore.Energy+2, attackerProdAfter.Energy,
		"Attacker energy production should increase by 2")
	testutil.AssertEqual(t, targetProdBefore.Plants-1, targetProdAfter.Plants,
		"Target plant production should decrease by 1")
}

// =============================================================================
// Card 166: Shuttles (active card with discount effect)
// "Decrease your energy production 1 step and increase your M$ production 2 steps."
// "Effect: When you play a space card, you pay 2 M$ less for it."
// =============================================================================

func TestShuttles_OnPlay_ProductionChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	shuttles := gamecards.Card{
		ID:   "card-shuttles",
		Name: "Shuttles",
		Type: gamecards.CardTypeActive,
		Pack: "base-game",
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{
						ResourceType: shared.ResourceDiscount,
						Amount:       2,
						Target:       "self-player",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagSpace}},
						},
					},
				},
			},
		},
	}

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{shuttles})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Oxygen needs to be 5 for the requirement
	testGame.GlobalParameters().IncreaseOxygen(ctx, 5)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-shuttles")

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-shuttles", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Shuttles should play successfully")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits, "Credit production should increase by 2")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy, "Energy production should decrease by 1")
}

// =============================================================================
// Card 199: Restricted Area (active card with action)
// "Action: Spend 2 M$ to draw a card."
// =============================================================================

func TestRestrictedArea_SpendCreditsToDrawCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-restricted-area"
	p.PlayedCards().AddCard(cardID, "Restricted Area", "active", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Restricted Area",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	handBefore := p.Hand().CardCount()

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Restricted Area action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 8, resources.Credits, "Should have 8 credits after spending 2")

	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore+1, handAfter, "Hand should increase by 1 from drawing 1 card")
}

func TestRestrictedArea_FailsWithoutEnoughCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-restricted-area"
	p.PlayedCards().AddCard(cardID, "Restricted Area", "active", []string{"science"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 1, // Only 1, need 2
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Restricted Area",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Restricted Area should fail without enough credits")
}
