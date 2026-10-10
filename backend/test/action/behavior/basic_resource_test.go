package behavior_test

import (
	"context"
	"log/slog"
	"testing"

	cardAction "openmars/internal/action/card"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
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

	payment := shared.NativePayment(shared.ResourceCredit, 12)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, &payment, nil, nil)
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
	payment := &shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-player",

		Resource: "credit"}, TargetResource: shared.ResourceCredit,
		Amount: 6}, {Source: shared.PaymentSource{Target: "self-player",
		Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 2}},
	}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil, nil)
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
	payment := &shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-player",

		Resource: "credit"}, TargetResource: shared.ResourceCredit,
		Amount: 3}, {Source: shared.PaymentSource{Target: "self-player",
		Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 1}},
	}
	before := p.Resources().Get()
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil, nil)
	testutil.AssertErrorContains(t, err, "insufficient credit payment", "Should fail with insufficient payment")
	testutil.AssertEqual(t, before, p.Resources().Get(), "rejected payment spends nothing")
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
	payment := &shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-player",

		Resource: "credit"}, TargetResource: shared.ResourceCredit,
		Amount: 6}, {Source: shared.PaymentSource{Target: "self-player",
		Resource: "steel"}, TargetResource: shared.ResourceCredit,
		Amount: 3}},
	}
	before := p.Resources().Get()
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil, nil)
	testutil.AssertErrorContains(t, err, "ineligible payment source", "Should fail when using steel (not allowed)")
	testutil.AssertEqual(t, before, p.Resources().Get(), "rejected payment spends nothing")
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
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil, nil, nil)
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
	payment := &shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-player",

		Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 2}},
	}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, nil, nil, nil, nil, payment, nil, nil)
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
	payment := &shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-player",

		Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 3}},
	}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment, nil, nil)
	testutil.AssertNoError(t, err, "Should succeed with titanium value modifier")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Should have 0 titanium after spending 3")
}

func TestBasicResource_DeferredRemovalWithAdjacentRestriction(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	target, _ := testGame.GetPlayer(targetPlayerID)
	target.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 10})

	removalOutput := &shared.BasicResourceCondition{
		ConditionBase:     shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: -3, Target: "any-player"},
		TargetRestriction: &shared.TargetRestriction{Adjacent: "self-card"},
	}

	applier := gamecards.NewBehaviorApplier(p, testGame, "test", slog.Default()).
		WithCardRegistry(cardRegistry).
		WithTargetPlayerID(targetPlayerID)
	err := applier.ApplyOutputs(context.Background(), []shared.BehaviorCondition{removalOutput})
	testutil.AssertNoError(t, err, "ApplyOutputs should succeed")

	testutil.AssertTrue(t, applier.DeferredRemoval() != nil, "Removal should be deferred, not applied immediately")
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

func TestRestrictedResourceRemoval_CombinesRestrictions(t *testing.T) {
	ctx := context.Background()
	g, repo, registry, id, otherID := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	target, _ := g.GetPlayer(otherID)
	target.Resources().Set(shared.Resources{Steel: 7})
	c := testutil.GetCardByName("Dirigibles")
	target.PlayedCards().AddCard(c.ID, c.Name, string(c.Type), []string{"venus"})
	output := shared.NewBasicResourceCondition(shared.ResourceSteel, -3, "any-player")
	output.TargetRestriction = &shared.TargetRestriction{Selectors: []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}}, Adjacent: "self-card"}
	position := shared.HexPosition{Q: 4, R: -1, S: -3}
	targets, err := gamecards.ResourceRemovalTargets(g, p, output, &position, registry)
	testutil.AssertNoError(t, err, "resolve tags and adjacency")
	testutil.AssertEqual(t, 0, len(targets), "tag alone is insufficient")
	testutil.PlaceTileForPlayer(ctx, t, g, repo, otherID, "city", "3,-1,-2")
	targets, err = gamecards.ResourceRemovalTargets(g, p, output, &position, registry)
	testutil.AssertNoError(t, err, "resolve matching target")
	testutil.AssertEqual(t, 3, targets[otherID], "generic resource type with both restrictions")
	output.TargetRestriction.Selectors = []shared.Selector{{Tags: []shared.CardTag{shared.TagEarth}}}
	targets, err = gamecards.ResourceRemovalTargets(g, p, output, &position, registry)
	testutil.AssertNoError(t, err, "resolve nonmatching tag")
	testutil.AssertEqual(t, 0, len(targets), "adjacency alone is insufficient")
}

func TestRestrictedResourceRemoval_UsesCardSelectors(t *testing.T) {
	testCards := []gamecards.Card{
		{ID: "venus-only", Type: gamecards.CardTypeActive, Tags: []shared.CardTag{shared.TagVenus}},
		{ID: "science-only", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}},
		{ID: "venus-science", Type: gamecards.CardTypeActive, Cost: 12, Tags: []shared.CardTag{shared.TagVenus, shared.TagScience}},
		{ID: "venus-event", Type: gamecards.CardTypeEvent, Tags: []shared.CardTag{shared.TagVenus}},
		{ID: "wild-only", Type: gamecards.CardTypePrelude, Tags: []shared.CardTag{shared.TagWild}},
	}
	minCost := 10
	for _, tc := range []struct {
		name      string
		cards     []string
		selectors []shared.Selector
		matches   bool
	}{
		{"tags on separate cards do not combine", []string{"venus-only", "science-only"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus, shared.TagScience}}}, false},
		{"tags on one card match", []string{"venus-science"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus, shared.TagScience}}}, true},
		{"selectors use OR", []string{"science-only"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}, {Tags: []shared.CardTag{shared.TagScience}}}, true},
		{"fields use AND", []string{"venus-only"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}, CardTypes: []string{"automated"}}}, false},
		{"cost constraint", []string{"venus-science"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}, RequiredOriginalCost: &shared.MinMaxValue{Min: &minCost}}}, true},
		{"cost constraint rejects", []string{"venus-only"}, []shared.Selector{{RequiredOriginalCost: &shared.MinMaxValue{Min: &minCost}}}, false},
		{"event excluded", []string{"venus-event"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}}, false},
		{"wild is not a Venus tag", []string{"wild-only"}, []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, _, id, otherID := testutil.SetupTwoPlayerGame(t)
			registry := testutil.CreateTestCardRegistryWithAdditionalCards(testCards)
			actor, _ := g.GetPlayer(id)
			target, _ := g.GetPlayer(otherID)
			target.Resources().Set(shared.Resources{Credits: 10})
			for _, id := range tc.cards {
				c, err := registry.GetByID(id)
				testutil.AssertNoError(t, err, "get card")
				target.PlayedCards().AddCard(c.ID, c.Name, string(c.Type), nil)
			}
			output := shared.NewBasicResourceCondition(shared.ResourceCredit, -4, "any-player")
			output.TargetRestriction = &shared.TargetRestriction{Selectors: tc.selectors}
			targets, err := gamecards.ResourceRemovalTargets(g, actor, output, nil, registry)
			testutil.AssertNoError(t, err, "evaluate selectors")
			testutil.AssertEqual(t, tc.matches, targets[otherID] > 0, "card selector eligibility")
		})
	}
}
