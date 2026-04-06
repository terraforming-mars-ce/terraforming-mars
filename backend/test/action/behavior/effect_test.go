package behavior_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Storage Payment Substitute (Dirigibles pattern) ---

func TestStoragePaymentSubstitute_RegisteredOnCardPlay(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Register a test card that has storage-payment-substitute auto behavior
	dirigiblesID := "test-dirigibles"
	testCards := []gamecards.Card{
		{
			ID:   dirigiblesID,
			Name: "Test Dirigibles",
			Type: gamecards.CardTypeActive,
			Cost: 11,
			Tags: []shared.CardTag{shared.TagVenus},
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Outputs: []shared.BehaviorCondition{
						&shared.EffectCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceStoragePaymentSubstitute, Amount: 3}, Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagVenus}},
						}},
					},
				},
			},
			ResourceStorage: &gamecards.ResourceStorage{
				Type:     shared.ResourceFloater,
				Starting: 0,
			},
		},
	}

	testCardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(testCards)
	p.Hand().AddCard(dirigiblesID)

	playAction := cardAction.NewPlayCardAction(repo, testCardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playAction.Execute(ctx, testGame.ID(), playerID, dirigiblesID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Dirigibles should succeed")

	// Verify storage payment substitute was registered
	subs := p.Resources().StoragePaymentSubstitutes()
	testutil.AssertTrue(t, len(subs) > 0, "Should have at least one storage payment substitute")
	testutil.AssertEqual(t, dirigiblesID, subs[0].CardID, "Storage payment substitute should reference Dirigibles card")
	testutil.AssertEqual(t, 3, subs[0].ConversionRate, "Floater conversion rate should be 3")
}

func TestStoragePaymentSubstitute_UsedForCardPayment(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Manually set up a storage payment substitute (simulating Dirigibles already played)
	dirigiblesID := "test-dirigibles-played"
	p.PlayedCards().AddCard(dirigiblesID, "Test Dirigibles", "active", []string{"venus"})
	p.Resources().AddToStorage(dirigiblesID, 3) // 3 floaters stored

	p.Resources().AddStoragePaymentSubstitute(shared.StoragePaymentSubstitute{
		CardID:         dirigiblesID,
		ResourceType:   shared.ResourceFloater,
		ConversionRate: 3,
		Selectors:      []shared.Selector{{Tags: []shared.CardTag{shared.TagVenus}}},
	})

	// Define a synthetic Venus-tagged card for this test
	venusCardID := "test-venus-card"
	syntheticVenusCard := gamecards.Card{
		ID:   venusCardID,
		Name: "Test Venus Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 7,
		Tags: []shared.CardTag{shared.TagVenus},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{syntheticVenusCard})

	p.Hand().AddCard(venusCardID)

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{
		Credits:            1, // 1 credit + 2 floaters * 3 M€ = 7 M€
		StorageSubstitutes: map[string]int{dirigiblesID: 2},
	}

	err := playAction.Execute(ctx, testGame.ID(), playerID, venusCardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Venus card with storage payment should succeed")

	// Verify floaters were deducted
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(dirigiblesID), "Should have 1 floater remaining after using 2")
}

// --- CardPayment unit tests ---

func TestCardPayment_TotalValue_WithStorageSubstitutes(t *testing.T) {
	substitutes := []shared.PaymentSubstitute{
		{ResourceType: shared.ResourceSteel, ConversionRate: 2},
		{ResourceType: shared.ResourceTitanium, ConversionRate: 3},
	}
	storageSubs := []shared.StoragePaymentSubstitute{
		{CardID: "card-a", ResourceType: shared.ResourceFloater, ConversionRate: 3},
	}

	payment := gamecards.CardPayment{
		Credits:            5,
		StorageSubstitutes: map[string]int{"card-a": 2},
	}

	total := payment.TotalValue(substitutes, storageSubs)
	// 5 credits + 2 floaters * 3 M€ = 11 M€
	testutil.AssertEqual(t, 11, total, "Total value should include storage substitute value")
}

func TestCardPayment_CoversCardCost_WithStorageSubstitutes(t *testing.T) {
	substitutes := []shared.PaymentSubstitute{
		{ResourceType: shared.ResourceSteel, ConversionRate: 2},
		{ResourceType: shared.ResourceTitanium, ConversionRate: 3},
	}
	storageSubs := []shared.StoragePaymentSubstitute{
		{CardID: "card-a", ResourceType: shared.ResourceFloater, ConversionRate: 3},
	}

	payment := gamecards.CardPayment{
		Credits:            5,
		StorageSubstitutes: map[string]int{"card-a": 2},
	}

	err := payment.CoversCardCost(11, false, false, substitutes, storageSubs)
	testutil.AssertNoError(t, err, "Payment of 5 credits + 2*3 floaters should cover 11 M€ cost")
}

func TestCardPayment_CoversCardCost_RejectsInvalidStorageCard(t *testing.T) {
	substitutes := []shared.PaymentSubstitute{
		{ResourceType: shared.ResourceSteel, ConversionRate: 2},
		{ResourceType: shared.ResourceTitanium, ConversionRate: 3},
	}

	// No storage substitutes registered
	payment := gamecards.CardPayment{
		Credits:            5,
		StorageSubstitutes: map[string]int{"unregistered-card": 2},
	}

	err := payment.CoversCardCost(11, false, false, substitutes, nil)
	testutil.AssertError(t, err, "Should reject payment from unregistered storage card")
}

func TestCardPayment_CanAfford_ChecksStorageAvailability(t *testing.T) {
	payment := gamecards.CardPayment{
		Credits:            5,
		StorageSubstitutes: map[string]int{"card-a": 3},
	}

	resources := shared.Resources{Credits: 10}
	storageGetter := func(cardID string) int {
		if cardID == "card-a" {
			return 2 // Only 2 available, but trying to use 3
		}
		return 0
	}

	err := payment.CanAfford(resources, storageGetter)
	testutil.AssertError(t, err, "Should fail when trying to use more storage resources than available")
}

// --- Variable Amount Storage Inputs (Sulphur-Eating Bacteria pattern) ---

func TestVariableAmount_StorageInput_SpendMultipleMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-sulphur-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 5) // 5 microbes stored

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceMicrobe, 1, "self-card"),
				},
			},
			{
				Inputs: []shared.BehaviorCondition{
					&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"}, VariableAmount: true},
				},
				Outputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player"}, VariableAmount: true},
				},
			},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Choice 1: spend 3 microbes, gain 9 credits (3 * 3 M€)
	choiceIndex := 1
	selectedAmount := 3

	creditsBefore := p.Resources().Get().Credits

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, nil, nil, nil, &selectedAmount, nil, nil)
	testutil.AssertNoError(t, err, "Spending 3 microbes should succeed")

	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Should have 2 microbes remaining after spending 3")
	testutil.AssertEqual(t, creditsBefore+9, p.Resources().Get().Credits, "Should gain 9 credits (3 * 3)")
}

func TestVariableAmount_StorageInput_InsufficientMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-sulphur-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 2) // Only 2 microbes

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceMicrobe, 1, "self-card"),
				},
			},
			{
				Inputs: []shared.BehaviorCondition{
					&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"}, VariableAmount: true},
				},
				Outputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player"}, VariableAmount: true},
				},
			},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Try to spend 5 microbes when only 2 available
	choiceIndex := 1
	selectedAmount := 5

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, nil, nil, nil, &selectedAmount, nil, nil)
	testutil.AssertError(t, err, "Should fail when trying to spend more microbes than available")

	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Microbes should not be deducted on failure")
}

func TestEffect_DiscountWithTagSelector(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player"},
		Selectors:     []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}},
	}
	p.Effects().AddEffect(shared.CardEffect{
		CardID:   "test-discount-card",
		CardName: "Test Discount Card",
		Behavior: shared.CardBehavior{Outputs: []shared.BehaviorCondition{effect}},
	})

	spaceCard := &gamecards.Card{
		ID:   "test-space-card",
		Name: "Test Space Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
	}
	nonSpaceCard := &gamecards.Card{
		ID:   "test-non-space-card",
		Name: "Test Non-Space Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding},
	}

	testCards := []gamecards.Card{*spaceCard, *nonSpaceCard}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(testCards)

	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)
	spaceDiscount := calc.CalculateCardDiscounts(p, spaceCard)
	nonSpaceDiscount := calc.CalculateCardDiscounts(p, nonSpaceCard)

	testutil.AssertEqual(t, 2, spaceDiscount, "Space-tagged card should get 2 credit discount")
	testutil.AssertEqual(t, 0, nonSpaceDiscount, "Non-space card should get no discount")
}

func TestEffect_PaymentSubstituteRegistration(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePaymentSubstitute, Amount: 1, Target: "self-player"},
		Selectors:     []shared.Selector{{Resources: []string{"heat"}}},
	}

	applyOutputs(t, p, testGame, cardRegistry, effect)

	subs := p.Resources().PaymentSubstitutes()
	found := false
	for _, sub := range subs {
		if sub.ResourceType == shared.ResourceHeat {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Payment substitutes should contain heat")
}

func TestEffect_ValueModifierApplication(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceValueModifier, Amount: 1, Target: "self-player"},
		Selectors:     []shared.Selector{{Resources: []string{"titanium"}}},
	}

	applyOutputs(t, p, testGame, cardRegistry, effect)

	modifier := p.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 1, modifier, "Titanium value modifier should be 1")
}

func TestEffect_GlobalParameterLenience(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGlobalParameterLenience, Amount: 2, Target: "self-player"},
		Selectors:     []shared.Selector{{GlobalParameters: []string{"temperature"}}},
	}
	p.Effects().AddEffect(shared.CardEffect{
		CardID:   "test-lenience-card",
		CardName: "Test Lenience Card",
		Behavior: shared.CardBehavior{Outputs: []shared.BehaviorCondition{effect}},
	})

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(nil)
	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)
	lenience := calc.CalculateGlobalParameterLenience(p, "temperature")

	testutil.AssertEqual(t, 2, lenience, "Temperature lenience should be 2")
}

func TestEffect_IgnoreGlobalRequirements(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceIgnoreGlobalRequirements, Amount: 1, Target: "self-player"},
	}
	p.Effects().AddEffect(shared.CardEffect{
		CardID:   "test-ignore-card",
		CardName: "Test Ignore Card",
		Behavior: shared.CardBehavior{Outputs: []shared.BehaviorCondition{effect}},
	})

	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(nil)
	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)
	hasIgnore := calc.HasIgnoreGlobalRequirements(p)

	testutil.AssertTrue(t, hasIgnore, "Player should have ignore global requirements")
}

func TestEffect_TemporaryNextCard(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceDiscount, Amount: 2, Target: "self-player"},
		Selectors:     []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}},
		Temporary:     "next-card",
	}

	cardEffect := shared.CardEffect{
		CardID:   "test-temporary-card",
		CardName: "Test Temporary Card",
		Behavior: shared.CardBehavior{
			Outputs: []shared.BehaviorCondition{effect},
		},
	}
	p.Effects().AddEffect(cardEffect)

	effects := p.Effects().List()
	testutil.AssertTrue(t, len(effects) > 0, "Player should have at least one effect")

	lastEffect := effects[len(effects)-1]
	testutil.AssertTrue(t, len(lastEffect.Behavior.Outputs) > 0, "Effect behavior should have outputs")

	temporary := shared.GetTemporary(lastEffect.Behavior.Outputs[0])
	testutil.AssertEqual(t, "next-card", temporary, "Temporary field should be next-card")
}

func TestEffect_PaymentSubstituteEmptySelectors(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	subsBefore := len(p.Resources().PaymentSubstitutes())

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePaymentSubstitute, Amount: 1, Target: "self-player"},
		Selectors:     nil,
	}

	applyOutputs(t, p, testGame, cardRegistry, effect)

	subsAfter := p.Resources().PaymentSubstitutes()
	testutil.AssertEqual(t, subsBefore, len(subsAfter),
		"No payment substitute should be registered when selectors are empty")
}

func TestEffect_ValueModifierEmptySelectors(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	effect := &shared.EffectCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceValueModifier, Amount: 1, Target: "self-player"},
		Selectors:     []shared.Selector{},
	}

	applyOutputs(t, p, testGame, cardRegistry, effect)

	// With empty selectors, GetResourcesFromSelectors returns empty, so the loop body never executes
	steelMod := p.Resources().GetValueModifier(shared.ResourceSteel)
	titaniumMod := p.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 0, steelMod, "Steel value modifier should be 0 with empty selectors")
	testutil.AssertEqual(t, 0, titaniumMod, "Titanium value modifier should be 0 with empty selectors")
}
