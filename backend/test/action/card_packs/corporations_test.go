package card_packs_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// =============================================================================
// Robinson Industries (PC3, corporation, prelude)
// "Action: Spend 4 M€ to increase (one of) your lowest production 1 step.
//
//	You start with 47 M€."
//
// Behavior 0: auto-corporation-start trigger, outputs credit 47
// Behavior 1: manual trigger, inputs credit 4, choices for +1 production
// =============================================================================

func TestRobinsonIndustries_ActionSucceedsWithSufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := testutil.CardID("Robinson Industries")
	p.SetCorporationID(cardID)
	p.PlayedCards().AddCard(cardID, "Robinson Industries", "corporation", []string{})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 4, Target: "self-player"}},
		},
		ChoicePolicy: &shared.ChoicePolicy{Type: shared.ChoicePolicyTypeLowest},
		Choices: []shared.Choice{
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceHeatProduction, Amount: 1, Target: "self-player"}},
			}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Robinson Industries",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Robinson Industries action should succeed with 10 credits")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 6, resources.Credits, "Should have 6 credits after spending 4")

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, 1, prodAfter.Credits, "Credit production should increase by 1")
}

func TestRobinsonIndustries_OnlyAllowsIncreasingLowestProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := testutil.CardID("Robinson Industries")
	p.SetCorporationID(cardID)
	p.PlayedCards().AddCard(cardID, "Robinson Industries", "corporation", []string{})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	// Set production so steel is at 3, everything else at 0
	// The "lowest" productions are credit, titanium, plant, energy, heat (all 0)
	// Steel at 3 should NOT be chooseable
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceSteelProduction: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 4, Target: "self-player"}},
		},
		ChoicePolicy: &shared.ChoicePolicy{Type: shared.ChoicePolicyTypeLowest},
		Choices: []shared.Choice{
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceSteelProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"}},
			}},
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceHeatProduction, Amount: 1, Target: "self-player"}},
			}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Robinson Industries",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	// Try to increase steel production (choice index 1) - steel is at 3, NOT the lowest
	// This should FAIL because Robinson Industries only allows increasing the LOWEST production
	steelChoice := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &steelChoice, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Robinson Industries should reject increasing steel production (3) when other productions are at 0")
}

func TestRobinsonIndustries_ActionFailsWithInsufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := testutil.CardID("Robinson Industries")
	p.SetCorporationID(cardID)
	p.PlayedCards().AddCard(cardID, "Robinson Industries", "corporation", []string{})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 4, Target: "self-player"}},
		},
		Choices: []shared.Choice{
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"}},
			}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Robinson Industries",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})

	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Robinson Industries action should fail with only 3 credits")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Credits, "Credits should remain unchanged at 3")
}

func TestRobinsonIndustries_StateCalculatorBlocksWhenUnaffordable(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 4, Target: "self-player"}},
		},
		Choices: []shared.Choice{
			{Outputs: []shared.BehaviorCondition{
				&shared.ProductionCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"}},
			}},
		},
	}

	cardID := testutil.CardID("Robinson Industries")

	_ = ctx
	state := action.CalculatePlayerCardActionState(cardID, behavior, 1, p, testGame)

	testutil.AssertFalse(t, state.Available(), "Action should be unavailable with only 3 credits")

	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientResources && err.Category == player.ErrorCategoryInput {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Should have insufficient-resources error")
}

// =============================================================================
// Saturn Systems (B11, corporation, cost 0, tags: [jovian])
// "Effect: Each time any Jovian tag is put into play, including this,
//
//	increase your M€ production 1 step. You start with 1 titanium production
//	and 42 M€."
//
// Behavior 0: auto-corporation-start trigger, outputs credit 42 + titanium-production 1
// Behavior 1: auto trigger with condition tag-played for jovian (target any-player),
//
//	outputs credit-production 1
//
// When played as a regular card via PlayCardAction:
//   - Behavior 0 (auto-corporation-start) is NOT processed by PlayCardAction
//   - Behavior 1 (auto with condition) is registered as a passive effect
//
// =============================================================================
func TestSaturnSystems_CorporationPlays(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Saturn Systems")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Saturn Systems should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Saturn Systems should be in played cards")
	// The auto-corporation-start behavior is NOT processed by PlayCardAction,
	// so credits and titanium production should be unchanged from the play action itself.
	// However, the conditional trigger (behavior 1) should be registered as a passive effect.
	effects := p.Effects().List()
	foundPassiveEffect := false
	for _, effect := range effects {
		if effect.CardID == card.ID && effect.CardName == "Saturn Systems" {
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
// Teractor (B12, corporation, cost 0, tags: [earth])
// "Effect: When playing an Earth card, you pay 3 M€ less for it.
//
//	You start with 60 M€."
//
// Behavior 0: auto-corporation-start trigger, outputs credit 60
// Behavior 1: auto trigger (no condition), outputs discount 3
//
//	with selector tags:[earth]
//
// When played as a regular card via PlayCardAction:
//   - Behavior 0 (auto-corporation-start) is NOT processed by PlayCardAction
//   - Behavior 1 (auto, no condition, persistent discount) is applied immediately
//     and registered as a persistent effect
//
// =============================================================================
func TestTeractor_CorporationPlays(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Teractor")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "UpdateStatus failed")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "UpdatePhase failed")
	err = testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testutil.AssertNoError(t, err, "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Teractor should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Teractor should be in played cards")
	// The auto behavior (behavior 1) with persistent discount output should be
	// registered as an effect. Verify the discount effect exists.
	effects := p.Effects().List()
	foundDiscountEffect := false
	for _, effect := range effects {
		if effect.CardID == card.ID && effect.CardName == "Teractor" {
			foundDiscountEffect = true
			testutil.AssertEqual(t, 1, effect.BehaviorIndex,
				"Discount effect should reference behavior index 1")
			// Verify the behavior contains the expected discount output
			testutil.AssertEqual(t, 1, len(effect.Behavior.Outputs),
				"Discount behavior should have 1 output")
			testutil.AssertEqual(t, shared.ResourceDiscount, effect.Behavior.Outputs[0].GetResourceType(),
				"Output should be a discount resource type")
			testutil.AssertEqual(t, 3, effect.Behavior.Outputs[0].GetAmount(),
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
