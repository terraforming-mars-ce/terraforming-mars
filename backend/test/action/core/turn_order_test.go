package core_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/confirmation"
	gameaction "terraforming-mars-backend/internal/action/game"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

func TestTurnOrderPreservedAfterProductionCardSelection(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Verify initial turn order: [player1, player2]
	testutil.AssertEqual(t, player1ID, testGame.TurnOrder()[0], "Player 1 should be first in initial turn order")
	testutil.AssertEqual(t, player2ID, testGame.TurnOrder()[1], "Player 2 should be second in initial turn order")

	// Both players pass to trigger production phase (turn order rotates to [player2, player1])
	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	err := skipAction.Execute(ctx, testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "Player 1 PASS should succeed")

	err = skipAction.Execute(ctx, testGame.ID(), player2ID)
	testutil.AssertNoError(t, err, "Player 2 PASS should succeed")

	// Verify: production phase triggered, turn order rotated
	testutil.AssertEqual(t, game.GamePhaseProductionAndCardDraw, testGame.CurrentPhase(), "Game should be in production phase")
	testutil.AssertEqual(t, player2ID, testGame.TurnOrder()[0], "Player 2 should be first after rotation")
	testutil.AssertEqual(t, player1ID, testGame.TurnOrder()[1], "Player 1 should be second after rotation")

	// Both players confirm production cards (select none)
	confirmAction := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, logger)

	err = confirmAction.Execute(ctx, testGame.ID(), player1ID, []string{})
	testutil.AssertNoError(t, err, "Player 1 confirm production cards should succeed")

	err = confirmAction.Execute(ctx, testGame.ID(), player2ID, []string{})
	testutil.AssertNoError(t, err, "Player 2 confirm production cards should succeed")

	// After all confirm, game transitions back to action phase
	testutil.AssertEqual(t, game.GamePhaseAction, testGame.CurrentPhase(), "Game should be back in action phase")

	// The current turn should match turnOrder[0] (player2, not a random player)
	testutil.AssertEqual(t, testGame.TurnOrder()[0], testGame.CurrentTurn().PlayerID(),
		"Current turn should match first player in turn order")
	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(),
		"Player 2 should have the first turn after production phase")
}
