package play_card_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Worms (130) ---
// "Requires 4% oxygen. Increase your plant production 1 step for every 2 microbe tags you have, including this."

// Helper to build the Worms card definition used across all tests
func makeWormsCard() gamecards.Card {
	selfPlayerTarget := "self-player"
	microbeTag := shared.TagMicrobe
	return gamecards.Card{
		ID:   "130",
		Name: "Worms",
		Type: gamecards.CardTypeAutomated,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.ProductionCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagMicrobe),
							Amount:       2,
							Target:       &selfPlayerTarget,
							Tag:          &microbeTag,
						},
					},
				},
			},
		},
	}
}

// 1 microbe tag before Worms → total 2 → floor(2/2)=1 → +1 plant production
func TestWorms_OneMicrobeTagBefore_GainsOnePlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard := gamecards.Card{
		ID: "card-microbe-1", Name: "Microbe Card", Type: gamecards.CardTypeAutomated,
		Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// 1 microbe tag already in play
	p.PlayedCards().AddCard("card-microbe-1", "Microbe Card", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 1 existing + 1 from Worms = 2 microbe tags; floor(2/2) * 1 = 1
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants,
		"Should gain +1 plant production (1 existing + 1 self = 2 microbe tags, floor(2/2)=1)")
}

// 0 microbe tags before Worms → total 1 (just Worms) → floor(1/2)=0 → +0 plant production
func TestWorms_ZeroMicrobeTagsBefore_GainsZeroPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// No microbe tags in play — Worms is the only microbe card
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// Only Worms itself = 1 microbe tag; floor(1/2) * 1 = 0
	testutil.AssertEqual(t, productionBefore.Plants, productionAfter.Plants,
		"Should gain +0 plant production (only 1 microbe tag from Worms, floor(1/2)=0)")
}

// 3 microbe tags before Worms → total 4 → floor(4/2)=2 → +2 plant production
func TestWorms_ThreeMicrobeTagsBefore_GainsTwoPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard1 := gamecards.Card{ID: "card-m1", Name: "Microbe 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard2 := gamecards.Card{ID: "card-m2", Name: "Microbe 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard3 := gamecards.Card{ID: "card-m3", Name: "Microbe 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard1, microbeCard2, microbeCard3})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.PlayedCards().AddCard("card-m1", "Microbe 1", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m2", "Microbe 2", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m3", "Microbe 3", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 3 existing + 1 from Worms = 4; floor(4/2) * 1 = 2
	testutil.AssertEqual(t, productionBefore.Plants+2, productionAfter.Plants,
		"Should gain +2 plant production (3 existing + 1 self = 4 microbe tags, floor(4/2)=2)")
}

// 2 microbe tags before Worms → total 3 → floor(3/2)=1 → +1 plant production (rounds down)
func TestWorms_TwoMicrobeTagsBefore_RoundsDown(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard1 := gamecards.Card{ID: "card-m1", Name: "Microbe 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard2 := gamecards.Card{ID: "card-m2", Name: "Microbe 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard1, microbeCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.PlayedCards().AddCard("card-m1", "Microbe 1", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m2", "Microbe 2", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 2 existing + 1 from Worms = 3; floor(3/2) * 1 = 1
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants,
		"Should gain +1 plant production (2 existing + 1 self = 3 microbe tags, floor(3/2)=1)")
}
