package play_card_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// createAllOpponentsDrawTestCard creates a card that draws 2 cards for self and 1 card for all opponents.
// Simplified version of Sponsored Academies (without the card-discard).
func createAllOpponentsDrawTestCard() gamecards.Card {
	return gamecards.Card{
		ID:   "card-all-opponents-draw-test",
		Name: "All Opponents Draw Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardOperationCondition(shared.ResourceCardDraw, 2, "self-player"),
					shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "all-opponents"),
				},
			},
		},
	}
}

func TestPlayCard_AllOpponentsDrawCard_TwoPlayers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player2.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player1HandBefore := player1.Hand().CardCount()
	player2HandBefore := player2.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card should succeed")

	// Player 1: played 1 card (-1) and drew 2 cards (+2) = net +1
	player1HandAfter := player1.Hand().CardCount()
	testutil.AssertEqual(t, player1HandBefore+1, player1HandAfter,
		"Player 1 hand should increase by 1 (played 1, drew 2)")

	// Player 2 (opponent): drew 1 card (+1)
	player2HandAfter := player2.Hand().CardCount()
	testutil.AssertEqual(t, player2HandBefore+1, player2HandAfter,
		"Player 2 (opponent) should have drawn 1 card")
}

func TestPlayCard_AllOpponentsDrawCard_ThreePlayers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]
	player3 := players[2]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player2.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player3.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player2HandBefore := player2.Hand().CardCount()
	player3HandBefore := player3.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card should succeed")

	// Both opponents should have drawn 1 card each
	player2HandAfter := player2.Hand().CardCount()
	testutil.AssertEqual(t, player2HandBefore+1, player2HandAfter,
		"Player 2 (opponent) should have drawn 1 card")

	player3HandAfter := player3.Hand().CardCount()
	testutil.AssertEqual(t, player3HandBefore+1, player3HandAfter,
		"Player 3 (opponent) should have drawn 1 card")
}

func TestPlayCard_AllOpponentsDrawCard_SoloMode(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player1HandBefore := player1.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card in solo mode should succeed")

	// Solo: played 1 card (-1) and drew 2 cards (+2), no opponents = net +1
	player1HandAfter := player1.Hand().CardCount()
	testutil.AssertEqual(t, player1HandBefore+1, player1HandAfter,
		"Player 1 hand should increase by 1 in solo mode (played 1, drew 2, no opponents)")
}
