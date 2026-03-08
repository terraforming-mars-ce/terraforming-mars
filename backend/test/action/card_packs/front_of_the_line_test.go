package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Front of the Line (EXP001) ---
// Event card, cost 3. Grants 2 extra actions on the current turn.

func TestFrontOfTheLine_GrantsExtraActions(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	// Player starts with 2 actions
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should start with 2 actions")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().TotalActions(), "Should start with 2 total actions")

	err := playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play successfully")

	// Started with 2, consumed 1, gained 2 → 3 remaining, 4 total
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().ActionsRemaining(), "Should have 3 actions remaining (2 - 1 + 2)")
	testutil.AssertEqual(t, 4, testGame.CurrentTurn().TotalActions(), "Should have 4 total actions (2 + 2)")
}

func TestFrontOfTheLine_AsLastAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Set to 1 action remaining
	err := testGame.SetCurrentTurn(ctx, playerID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play as last action")

	// Started with 1, consumed 1, gained 2 → 2 remaining, 3 total
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should have 2 actions remaining (1 - 1 + 2)")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().TotalActions(), "Should have 3 total actions (1 + 2)")
	// Turn should NOT auto-advance since actions remain
	testutil.AssertEqual(t, playerID, testGame.CurrentTurn().PlayerID(), "Turn should still be current player's")
}

func TestFrontOfTheLine_UnlimitedActionsUnchanged(t *testing.T) {
	testGame, repo, cardRegistry, playerID := testutil.SetupSoloGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	// Verify unlimited actions
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should have unlimited actions")

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err := playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play in solo mode")

	// Unlimited actions should remain unchanged
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should still have unlimited actions")
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().TotalActions(), "Solo total actions should still be unlimited")
}

func TestFrontOfTheLine_ExtraActionsCanBeUsed(t *testing.T) {
	testGame, repo, cardRegistry, playerID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 200)

	// Start with 1 action so playing Front of the Line + 2 more cards uses all actions
	err := testGame.SetCurrentTurn(ctx, playerID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	// Play Front of the Line as the last action → gains 2 extra
	p.Hand().AddCard(testutil.CardID("Front of the Line"))
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play successfully")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should have 2 actions after Front of the Line")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().TotalActions(), "Should have 3 total actions (1 + 2)")

	// Use first extra action: play Power Plant (cost 4)
	p.Hand().AddCard(testutil.CardID("Power Plant"))
	payment2 := cardAction.PaymentRequest{Credits: 4}
	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Power Plant"), payment2, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "First extra action should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().ActionsRemaining(), "Should have 1 action remaining")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().TotalActions(), "Total actions should still be 3")
	testutil.AssertEqual(t, playerID, testGame.CurrentTurn().PlayerID(), "Should still be current player's turn")

	// Use second extra action: play Asteroid (cost 14)
	p.Hand().AddCard(testutil.CardID("Asteroid"))
	payment3 := cardAction.PaymentRequest{Credits: 14}
	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Asteroid"), payment3, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Second extra action should succeed")

	// All actions consumed → turn should auto-advance to player 2
	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(), "Turn should advance to player 2")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player 2 should have 2 actions")
}

// TestFrontOfTheLine_CardIsEvent verifies the card is an event type
func TestFrontOfTheLine_CardIsEvent(t *testing.T) {
	card := testutil.GetCardByName("Front of the Line")
	testutil.AssertEqual(t, "event", string(card.Type), "Front of the Line should be an event card")
	testutil.AssertEqual(t, 3, card.Cost, "Front of the Line should cost 3")

	// Verify it has the extra-actions output
	found := false
	for _, behavior := range card.Behaviors {
		for _, output := range behavior.Outputs {
			if output.ResourceType == shared.ResourceExtraActions && output.Amount == 2 {
				found = true
			}
		}
	}
	testutil.AssertTrue(t, found, "Front of the Line should have extra-actions output of 2")
}
