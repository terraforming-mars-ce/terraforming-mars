package behavior_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Water Import From Europa (012) ---
// "Action: Pay 12 M€ to place an ocean tile. Titanium may be used as if playing a space card."

func TestWaterImportFromEuropa_PayWithCreditsOnly(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	payment := &gamecards.CardPayment{Credits: 12}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil)
	testutil.AssertNoError(t, err, "Water Import action should succeed with credits only")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 8, resources.Credits, "Should have 8 credits after paying 12")
}

func TestWaterImportFromEuropa_PayWithTitaniumAndCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   6,
		shared.ResourceTitanium: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay 6 credits + 2 titanium (value 3 each = 6) = 12 total
	payment := &gamecards.CardPayment{Credits: 6, Titanium: 2}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil)
	testutil.AssertNoError(t, err, "Water Import action should succeed with titanium + credits")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Credits, "Should have 0 credits after paying 6")
	testutil.AssertEqual(t, 1, resources.Titanium, "Should have 1 titanium after spending 2")
}

func TestWaterImportFromEuropa_FailInsufficientPayment(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   3,
		shared.ResourceTitanium: 1,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay 3 credits + 1 titanium (value 3) = 6, need 12
	payment := &gamecards.CardPayment{Credits: 3, Titanium: 1}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil)
	testutil.AssertError(t, err, "Should fail with insufficient payment")
}

func TestWaterImportFromEuropa_FailSteelNotAllowed(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 6,
		shared.ResourceSteel:  5,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Try to pay with steel (not allowed)
	payment := &gamecards.CardPayment{Credits: 6, Steel: 3}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil)
	testutil.AssertError(t, err, "Should fail when using steel (not allowed)")
}

func TestWaterImportFromEuropa_NoPaymentFallsBackToCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 15,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// No payment provided — should fall back to credits-only
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should succeed with no payment (falls back to credits)")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Credits, "Should have 3 credits after paying 12")
}

// --- Rotator Impacts (243) ---
// Choice 1: "Spend 6 M€ to add an asteroid resource (titanium may be used)"

func TestRotatorImpacts_Choice1_PayWithTitanium(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "243"
	p.PlayedCards().AddCard(cardID, "Rotator Impacts", "active", []string{"space"})
	p.Resources().AddToStorage(cardID, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   0,
		shared.ResourceTitanium: 2,
	})

	choiceIndex := 0
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 6, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
				},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceAsteroid, 1, "self-card"),
				},
			},
			{
				Inputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceAsteroid, 1, "self-card"),
				},
				Outputs: []shared.BehaviorCondition{
					shared.NewGlobalParameterCondition(shared.ResourceVenus, 1, "none"),
				},
			},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Rotator Impacts",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay with 2 titanium (3 MC each = 6 MC total)
	payment := &gamecards.CardPayment{Credits: 0, Titanium: 2}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, nil, nil, nil, nil, payment, nil)
	testutil.AssertNoError(t, err, "Rotator Impacts choice 1 should succeed with titanium")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Should have 0 titanium after spending 2")
}

func TestWaterImportFromEuropa_TitaniumWithValueModifier(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	// Add titanium value modifier (+1, so titanium = 4 MC each)
	p.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   0,
		shared.ResourceTitanium: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 12, Target: "self-player"}, PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
		},
		Outputs: []shared.BehaviorCondition{
			shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay with 3 titanium (4 MC each with modifier = 12 MC total)
	payment := &gamecards.CardPayment{Credits: 0, Titanium: 3}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil)
	testutil.AssertNoError(t, err, "Should succeed with titanium value modifier")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Should have 0 titanium after spending 3")
}

func TestBasicResource_DeferredStealWithAdjacentRestriction(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	target, _ := testGame.GetPlayer(targetPlayerID)
	target.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 10})

	stealOutput := &shared.BasicResourceCondition{
		ConditionBase:     shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 3, Target: "steal-any-player"},
		TargetRestriction: &shared.TargetRestriction{Adjacent: "self-card"},
	}

	applier := gamecards.NewBehaviorApplier(p, testGame, "test", testutil.TestLogger()).
		WithCardRegistry(cardRegistry).
		WithTargetPlayerID(targetPlayerID)
	err := applier.ApplyOutputs(context.Background(), []shared.BehaviorCondition{stealOutput})
	testutil.AssertNoError(t, err, "ApplyOutputs should succeed")

	testutil.AssertTrue(t, applier.DeferredSteal() != nil, "Steal should be deferred, not applied immediately")
	testutil.AssertEqual(t, 10, target.Resources().Get().Credits, "Target credits should be unchanged")
}

func TestBasicResource_StealAnyPlayer(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	target, _ := testGame.GetPlayer(targetPlayerID)
	target.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 10})

	stealOutput := shared.NewBasicResourceCondition(shared.ResourceCredit, 3, "steal-any-player")

	creditsBefore := p.Resources().Get().Credits
	applyOutputsWithOptions(t, p, testGame, applyOptions{targetPlayerID: targetPlayerID, cardRegistry: cardRegistry}, stealOutput)

	testutil.AssertEqual(t, creditsBefore+3, p.Resources().Get().Credits, "Self should gain 3 credits")
	testutil.AssertEqual(t, 7, target.Resources().Get().Credits, "Target should lose 3 credits")
}

func TestBasicResource_AnyPlayerRemoval(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	target, _ := testGame.GetPlayer(targetPlayerID)
	target.Resources().Add(map[shared.ResourceType]int{shared.ResourcePlant: 5})

	removeOutput := shared.NewBasicResourceCondition(shared.ResourcePlant, -8, "any-player")

	plantsBefore := p.Resources().Get().Plants
	applyOutputsWithOptions(t, p, testGame, applyOptions{targetPlayerID: targetPlayerID, cardRegistry: cardRegistry}, removeOutput)

	testutil.AssertEqual(t, 0, target.Resources().Get().Plants, "Target plants should be clamped to 0")
	testutil.AssertEqual(t, plantsBefore, p.Resources().Get().Plants, "Self plants should be unchanged")
}

func TestBasicResource_AllSixTypes(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	outputs := []shared.BehaviorCondition{
		shared.NewBasicResourceCondition(shared.ResourceCredit, 5, "self-player"),
		shared.NewBasicResourceCondition(shared.ResourceSteel, 3, "self-player"),
		shared.NewBasicResourceCondition(shared.ResourceTitanium, 2, "self-player"),
		shared.NewBasicResourceCondition(shared.ResourcePlant, 4, "self-player"),
		shared.NewBasicResourceCondition(shared.ResourceEnergy, 1, "self-player"),
		shared.NewBasicResourceCondition(shared.ResourceHeat, 6, "self-player"),
	}

	applyOutputs(t, p, testGame, cardRegistry, outputs...)

	assertResources(t, p, map[shared.ResourceType]int{
		shared.ResourceCredit:   5,
		shared.ResourceSteel:    3,
		shared.ResourceTitanium: 2,
		shared.ResourcePlant:    4,
		shared.ResourceEnergy:   1,
		shared.ResourceHeat:     6,
	})
}

func TestBasicResource_VariableAmount(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	// Player starts with 0 credits (default)

	output := &shared.BasicResourceCondition{
		ConditionBase:  shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
		VariableAmount: true,
	}

	applyOutputsWithOptions(t, p, testGame, applyOptions{selectedAmount: 3, cardRegistry: cardRegistry}, output)

	assertResources(t, p, map[shared.ResourceType]int{
		shared.ResourceCredit: 3,
	})
}

func TestBasicResource_StealClampedToAvailable(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	target, _ := testGame.GetPlayer(targetPlayerID)
	target.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 2})

	stealOutput := shared.NewBasicResourceCondition(shared.ResourceCredit, 5, "steal-any-player")

	creditsBefore := p.Resources().Get().Credits
	applyOutputsWithOptions(t, p, testGame, applyOptions{targetPlayerID: targetPlayerID, cardRegistry: cardRegistry}, stealOutput)

	testutil.AssertEqual(t, creditsBefore+2, p.Resources().Get().Credits, "Self should gain only 2 (clamped to target's available)")
	testutil.AssertEqual(t, 0, target.Resources().Get().Credits, "Target should have 0 credits")
}

func TestBasicResource_StealSoloModeSkips(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 10})

	stealOutput := shared.NewBasicResourceCondition(shared.ResourceCredit, 3, "steal-any-player")

	applyOutputsWithOptions(t, p, testGame, applyOptions{cardRegistry: cardRegistry}, stealOutput)

	testutil.AssertEqual(t, 10, p.Resources().Get().Credits, "Self credits should be unchanged when no target player")
}

func TestBasicResource_ZeroAmountOutput(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 5})

	output := shared.NewBasicResourceCondition(shared.ResourceCredit, 0, "self-player")
	applyOutputs(t, p, testGame, cardRegistry, output)

	assertResources(t, p, map[shared.ResourceType]int{
		shared.ResourceCredit: 5,
	})
}

func TestBasicResource_VariableAmountZero(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	output := &shared.BasicResourceCondition{
		ConditionBase:  shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
		VariableAmount: true,
	}

	applyOutputsWithOptions(t, p, testGame, applyOptions{selectedAmount: 0, cardRegistry: cardRegistry}, output)

	assertResources(t, p, map[shared.ResourceType]int{
		shared.ResourceCredit: 0,
	})
}
