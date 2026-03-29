package card_packs_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/confirmation"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// =============================================================================
// Polyphemos (CC3, corporation, colonies)
// "Effect: Pay 5 M€ instead of 3 M€ when buying cards, including the starting hand.
//  You start with 50 M€. Increase your M€ production 5 steps. Gain 5 titanium."
// =============================================================================

func TestPolyphemos_NegativeDiscountOnCardBuying(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	polyphemosCard, err := cardRegistry.GetByID(testutil.CardID("Polyphemos"))
	testutil.AssertNoError(t, err, "Polyphemos card should exist")

	for behaviorIndex, behavior := range polyphemosCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        polyphemosCard.ID,
				CardName:      polyphemosCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	effects := p.Effects().List()
	testutil.AssertEqual(t, 1, len(effects), "Should have 1 discount effect registered")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discounts := calculator.CalculateActionDiscounts(p, shared.ActionCardBuying)
	testutil.AssertEqual(t, -2, discounts[shared.ResourceCredit], "Polyphemos should have -2 card-buying discount (cost increase)")
}

func TestPolyphemos_DoesNotAffectCardPlayCost(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	polyphemosCard, err := cardRegistry.GetByID(testutil.CardID("Polyphemos"))
	testutil.AssertNoError(t, err, "Polyphemos card should exist")

	for behaviorIndex, behavior := range polyphemosCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        polyphemosCard.ID,
				CardName:      polyphemosCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Polyphemos's -2 card-buying discount should NOT affect the play cost of cards
	anyCard, err := cardRegistry.GetByID(testutil.CardID("Arctic Algae"))
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	playDiscount := calculator.CalculateCardDiscounts(p, anyCard)
	testutil.AssertEqual(t, 0, playDiscount, "Polyphemos should not affect card play discount")
}

func TestPolyphemos_CardBuyCostFromCorpBehaviors(t *testing.T) {
	cardRegistry := testutil.CreateTestCardRegistry()
	polyphemosCard, err := cardRegistry.GetByID(testutil.CardID("Polyphemos"))
	testutil.AssertNoError(t, err, "Polyphemos card should exist")

	discount := gamecards.CalculateActionDiscountsFromCard(polyphemosCard, shared.ActionCardBuying)
	testutil.AssertEqual(t, -2, discount, "Polyphemos should have -2 card-buying discount from behaviors")

	effectiveCost := max(3-discount, 0)
	testutil.AssertEqual(t, 5, effectiveCost, "Polyphemos effective card buy cost should be 5 MC")
}

func TestPolyphemos_ProductionPhaseCardBuyCost(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	// Register Polyphemos discount effect
	polyphemosCard, err := cardRegistry.GetByID(testutil.CardID("Polyphemos"))
	testutil.AssertNoError(t, err, "Polyphemos card should exist")
	p.SetCorporationID(polyphemosCard.ID)

	for behaviorIndex, behavior := range polyphemosCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        polyphemosCard.ID,
				CardName:      polyphemosCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Set up production phase with available cards
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw), "update phase")
	drawnCards, err := testGame.Deck().DrawProjectCards(ctx, 4)
	testutil.AssertNoError(t, err, "draw project cards")

	testutil.AssertNoError(t, testGame.SetProductionPhase(ctx, playerID, &shared.ProductionPhase{
		AvailableCards:    drawnCards,
		SelectionComplete: false,
	}), "set production phase")

	// Give player exactly 10 credits (enough for 2 cards at 5 MC each, but not 3)
	testutil.SetPlayerCredits(ctx, p, 10)

	// Buying 2 cards should succeed (2 * 5 = 10 MC)
	action := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, nil, logger)
	err = action.Execute(ctx, testGame.ID(), playerID, drawnCards[:2])
	testutil.AssertNoError(t, err, "buying 2 cards at 5 MC each should succeed with 10 credits")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Credits, "should have 0 credits after buying 2 cards at 5 MC")
}

func TestPolyphemos_ProductionPhaseCardBuyCost_InsufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	// Register Polyphemos discount effect
	polyphemosCard, err := cardRegistry.GetByID(testutil.CardID("Polyphemos"))
	testutil.AssertNoError(t, err, "Polyphemos card should exist")
	p.SetCorporationID(polyphemosCard.ID)

	for behaviorIndex, behavior := range polyphemosCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        polyphemosCard.ID,
				CardName:      polyphemosCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Set up production phase
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw), "update phase")
	drawnCards, err := testGame.Deck().DrawProjectCards(ctx, 4)
	testutil.AssertNoError(t, err, "draw project cards")

	testutil.AssertNoError(t, testGame.SetProductionPhase(ctx, playerID, &shared.ProductionPhase{
		AvailableCards:    drawnCards,
		SelectionComplete: false,
	}), "set production phase")

	// Give player 9 credits (not enough for 2 cards at 5 MC each)
	testutil.SetPlayerCredits(ctx, p, 9)

	// Buying 2 cards should fail (2 * 5 = 10 MC, only have 9)
	action := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, nil, logger)
	err = action.Execute(ctx, testGame.ID(), playerID, drawnCards[:2])
	testutil.AssertError(t, err, "buying 2 cards at 5 MC each should fail with only 9 credits")
}

// =============================================================================
// Rim Freighters (C35, active, colonies)
// "Effect: Pay 1 less resource when trading."
// =============================================================================

func TestRimFreighters_ColonyTradeDiscount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	rimFreightersCard, err := cardRegistry.GetByID(testutil.CardID("Rim Freighters"))
	testutil.AssertNoError(t, err, "Rim Freighters card should exist")

	for behaviorIndex, behavior := range rimFreightersCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        rimFreightersCard.ID,
				CardName:      rimFreightersCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discounts := calculator.CalculateActionDiscounts(p, shared.ActionColonyTrade)

	testutil.AssertEqual(t, 1, discounts[shared.ResourceCredit], "Rim Freighters should give 1 credit trade discount")
	testutil.AssertEqual(t, 1, discounts[shared.ResourceEnergy], "Rim Freighters should give 1 energy trade discount")
	testutil.AssertEqual(t, 1, discounts[shared.ResourceTitanium], "Rim Freighters should give 1 titanium trade discount")
}

func TestRimFreighters_DoesNotAffectCardPlayCost(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	rimFreightersCard, err := cardRegistry.GetByID(testutil.CardID("Rim Freighters"))
	testutil.AssertNoError(t, err, "Rim Freighters card should exist")

	for behaviorIndex, behavior := range rimFreightersCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        rimFreightersCard.ID,
				CardName:      rimFreightersCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	anyCard, err := cardRegistry.GetByID(testutil.CardID("Arctic Algae"))
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	playDiscount := calculator.CalculateCardDiscounts(p, anyCard)
	testutil.AssertEqual(t, 0, playDiscount, "Rim Freighters should not affect card play discount")
}

func TestRimFreighters_DoesNotAffectCardBuyingCost(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	rimFreightersCard, err := cardRegistry.GetByID(testutil.CardID("Rim Freighters"))
	testutil.AssertNoError(t, err, "Rim Freighters card should exist")

	for behaviorIndex, behavior := range rimFreightersCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(shared.CardEffect{
				CardID:        rimFreightersCard.ID,
				CardName:      rimFreightersCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discounts := calculator.CalculateActionDiscounts(p, shared.ActionCardBuying)
	testutil.AssertEqual(t, 0, discounts[shared.ResourceCredit], "Rim Freighters should not affect card-buying cost")
}

func TestPartialResourceTradeDiscount_OnlyAffectsSpecifiedResources(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()
	p := players[0]

	// Register a fake effect that discounts only titanium and credit for colony-trade (not energy)
	p.Effects().AddEffect(shared.CardEffect{
		CardID:        "fake-partial-trade-discount",
		CardName:      "Fake Partial Trade Discount",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
			Outputs: []shared.BehaviorCondition{
				&shared.EffectCondition{
					ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceDiscount, Amount: 1, Target: "self-player"},
					Selectors: []shared.Selector{
						{
							Actions:   []string{shared.ActionColonyTrade},
							Resources: []string{"titanium", "credit"},
						},
					},
				},
			},
		},
	})

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discounts := calculator.CalculateActionDiscounts(p, shared.ActionColonyTrade)

	testutil.AssertEqual(t, 1, discounts[shared.ResourceTitanium], "Should discount titanium trade cost")
	testutil.AssertEqual(t, 1, discounts[shared.ResourceCredit], "Should discount credit trade cost")
	testutil.AssertEqual(t, 0, discounts[shared.ResourceEnergy], "Should NOT discount energy trade cost")
}
