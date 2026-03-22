package core_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/confirmation"
	gameaction "terraforming-mars-backend/internal/action/game"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game/shared"
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
	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	err := skipAction.Execute(ctx, testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "Player 1 PASS should succeed")

	err = skipAction.Execute(ctx, testGame.ID(), player2ID)
	testutil.AssertNoError(t, err, "Player 2 PASS should succeed")

	// Verify: production phase triggered, turn order rotated
	testutil.AssertEqual(t, shared.GamePhaseProductionAndCardDraw, testGame.CurrentPhase(), "Game should be in production phase")
	testutil.AssertEqual(t, player2ID, testGame.TurnOrder()[0], "Player 2 should be first after rotation")
	testutil.AssertEqual(t, player1ID, testGame.TurnOrder()[1], "Player 1 should be second after rotation")

	// Both players confirm production cards (select none)
	confirmAction := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, nil, logger)

	err = confirmAction.Execute(ctx, testGame.ID(), player1ID, []string{})
	testutil.AssertNoError(t, err, "Player 1 confirm production cards should succeed")

	err = confirmAction.Execute(ctx, testGame.ID(), player2ID, []string{})
	testutil.AssertNoError(t, err, "Player 2 confirm production cards should succeed")

	// After all confirm, game transitions back to action phase
	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Game should be back in action phase")

	// The current turn should match turnOrder[0] (player2, not a random player)
	testutil.AssertEqual(t, testGame.TurnOrder()[0], testGame.CurrentTurn().PlayerID(),
		"Current turn should match first player in turn order")
	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(),
		"Player 2 should have the first turn after production phase")
}

func TestTurnOrderRotatesCorrectlyWithFourPlayers(t *testing.T) {
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 4)
	logger := testutil.TestLogger()
	ctx := context.Background()

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmAction := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, nil, logger)

	// Verify initial turn order: [p0, p1, p2, p3]
	for i, id := range playerIDs {
		testutil.AssertEqual(t, id, testGame.TurnOrder()[i], "Initial turn order mismatch at position %d")
	}

	// Run through 4 generations to verify turn order rotates fully around
	for gen := 0; gen < 4; gen++ {
		// Expected first player rotates each generation: gen0=p0, gen1=p1, gen2=p2, gen3=p3
		expectedFirst := playerIDs[gen%len(playerIDs)]
		testutil.AssertEqual(t, expectedFirst, testGame.TurnOrder()[0],
			"Generation %d: wrong first player in turn order")
		testutil.AssertEqual(t, expectedFirst, testGame.CurrentTurn().PlayerID(),
			"Generation %d: current turn should match first player in turn order")

		// Verify full rotation order
		for i := range playerIDs {
			expectedPlayer := playerIDs[(gen+i)%len(playerIDs)]
			testutil.AssertEqual(t, expectedPlayer, testGame.TurnOrder()[i],
				"Generation %d: wrong player at position %d")
		}

		// All players pass (in current turn order)
		for _, id := range testGame.TurnOrder() {
			err := skipAction.Execute(ctx, testGame.ID(), id)
			testutil.AssertNoError(t, err, "Player PASS should succeed in generation %d")
		}

		testutil.AssertEqual(t, shared.GamePhaseProductionAndCardDraw, testGame.CurrentPhase(),
			"Generation %d: game should be in production phase")

		// All players confirm production cards
		for _, id := range playerIDs {
			err := confirmAction.Execute(ctx, testGame.ID(), id, []string{})
			testutil.AssertNoError(t, err, "Player confirm should succeed in generation %d")
		}

		testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(),
			"Generation %d: game should return to action phase")
	}

	// After 4 rotations with 4 players, turn order should be back to the original
	for i, id := range playerIDs {
		testutil.AssertEqual(t, id, testGame.TurnOrder()[i],
			"After full rotation cycle: turn order should match original at position %d")
	}
}

func TestTurnOrderRotatesCorrectlyWithThreePlayers(t *testing.T) {
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	logger := testutil.TestLogger()
	ctx := context.Background()

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmAction := confirmation.NewConfirmProductionCardsAction(repo, cardRegistry, nil, logger)

	// Run through 3 generations to verify full rotation cycle
	for gen := 0; gen < 3; gen++ {
		expectedFirst := playerIDs[gen%len(playerIDs)]
		testutil.AssertEqual(t, expectedFirst, testGame.TurnOrder()[0],
			"Generation %d: wrong first player in turn order")
		testutil.AssertEqual(t, expectedFirst, testGame.CurrentTurn().PlayerID(),
			"Generation %d: current turn should match first player in turn order")

		// All players pass
		for _, id := range testGame.TurnOrder() {
			err := skipAction.Execute(ctx, testGame.ID(), id)
			testutil.AssertNoError(t, err, "Player PASS should succeed in generation %d")
		}

		// All players confirm production cards
		for _, id := range playerIDs {
			err := confirmAction.Execute(ctx, testGame.ID(), id, []string{})
			testutil.AssertNoError(t, err, "Player confirm should succeed in generation %d")
		}

		testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(),
			"Generation %d: game should return to action phase")
	}

	// After 3 rotations with 3 players, turn order should be back to the original
	for i, id := range playerIDs {
		testutil.AssertEqual(t, id, testGame.TurnOrder()[i],
			"After full rotation cycle: turn order should match original at position %d")
	}
}
