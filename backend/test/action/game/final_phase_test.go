package game_test

import (
	"context"
	"testing"

	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	gameaction "terraforming-mars-backend/internal/action/game"
	resconv "terraforming-mars-backend/internal/action/resource_conversion"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestFinalPhase_TransitionAfterProduction verifies that after confirming
// final production when a player has enough plants, game enters final_phase phase.
func TestFinalPhase_TransitionAfterProduction(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// Give P1 enough plants for greenery (8 plants)
	p1.Resources().Set(shared.Resources{Plants: 16, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	// Both players pass → triggers final production
	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")

	testutil.AssertEqual(t, shared.GamePhaseProductionAndCardDraw, g.CurrentPhase(), "Should be in production phase")

	// Confirm production for both
	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm production")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm production")

	// P1 has 16 plants (enough for greenery), P2 has 0 → game should be in final greenery
	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(),
		"Game should be in final phase when a player has enough plants")

	// P2 should be auto-passed (0 plants < 8 required)
	p2After, _ := g.GetPlayer(p2ID)
	testutil.AssertTrue(t, p2After.HasPassed(), "P2 should be auto-passed (no plants)")

	// P1 should NOT be passed
	p1After, _ := g.GetPlayer(p1ID)
	testutil.AssertFalse(t, p1After.HasPassed(), "P1 should not be auto-passed (has plants)")

	// P1 should be the current turn player
	currentTurn := g.CurrentTurn()
	testutil.AssertTrue(t, currentTurn != nil, "Current turn should be set")
	testutil.AssertEqual(t, p1ID, currentTurn.PlayerID(), "P1 should have the current turn")
}

// TestFinalPhase_AllAutoPassGoesToScoring verifies that when no player
// has enough plants, the game goes straight to final scoring.
func TestFinalPhase_AllAutoPassGoesToScoring(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// No plants for either player
	p1.Resources().Set(shared.Resources{Plants: 0, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")

	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm production")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm production")

	// Both have 0 plants → all auto-pass → final scoring → game complete
	testutil.AssertEqual(t, shared.GameStatusCompleted, g.Status(),
		"Game should be completed when all players auto-pass greenery phase")
	testutil.AssertEqual(t, shared.GamePhaseComplete, g.CurrentPhase(),
		"Game phase should be complete")
}

// TestFinalPhase_PassTriggersScoring verifies that when the last player
// in the greenery phase passes, final scoring is triggered.
func TestFinalPhase_PassTriggersScoring(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// P1 has plants but will choose to pass
	p1.Resources().Set(shared.Resources{Plants: 16, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	// Pass through action phase → production → confirm
	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")

	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm")

	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(), "Should be in final phase")

	// P1 is last active player (P2 auto-passed), P1 passes
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass in final phase")

	testutil.AssertEqual(t, shared.GameStatusCompleted, g.Status(),
		"Game should be completed after all pass in final phase")
}

// TestFinalPhase_ConvertPlantsWorks verifies that converting plants
// to greenery works during the final phase.
func TestFinalPhase_ConvertPlantsWorks(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// P1 has 16 plants (enough for 2 greeneries)
	p1.Resources().Set(shared.Resources{Plants: 16, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)
	convertAction := resconv.NewConvertPlantsToGreeneryAction(repo, cardRegistry, nil, logger)

	// Pass through action phase → production → confirm
	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")
	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm")

	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(), "Should be in final phase")

	// P1 converts plants to greenery
	err = convertAction.Execute(ctx, g.ID(), p1ID, nil)
	testutil.AssertNoError(t, err, "P1 convert plants to greenery")

	// P1 should now have 8 plants remaining (16 - 8)
	p1After, _ := g.GetPlayer(p1ID)
	testutil.AssertEqual(t, 8, p1After.Resources().Get().Plants,
		"P1 should have 8 plants after converting once")

	// Game should still be in final phase
	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(),
		"Game should still be in final phase after one conversion")
}

// TestFinalPhase_TwoActionsPerTurn verifies that players get 2 actions
// per turn in the final phase.
func TestFinalPhase_TwoActionsPerTurn(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(playerIDs[0])
	p2, _ := g.GetPlayer(playerIDs[1])
	p3, _ := g.GetPlayer(playerIDs[2])

	// P1 and P2 have plants, P3 doesn't
	p1.Resources().Set(shared.Resources{Plants: 24, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 16, Credits: 10})
	p3.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	// All pass
	err := g.SetCurrentTurn(ctx, playerIDs[0], 2)
	testutil.AssertNoError(t, err, "Set current turn")
	for _, pID := range playerIDs {
		err = skipAction.Execute(ctx, g.ID(), pID)
		testutil.AssertNoError(t, err, "Player pass")
	}

	// All confirm production
	for _, pID := range playerIDs {
		err = confirmProdAction.Execute(ctx, g.ID(), pID, []string{})
		testutil.AssertNoError(t, err, "Player confirm production")
	}

	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(), "Should be in final phase")

	// P3 should be auto-passed
	p3After, _ := g.GetPlayer(playerIDs[2])
	testutil.AssertTrue(t, p3After.HasPassed(), "P3 should be auto-passed")

	// P1 should have 2 actions
	currentTurn := g.CurrentTurn()
	testutil.AssertEqual(t, playerIDs[0], currentTurn.PlayerID(), "P1 should have current turn")
	testutil.AssertEqual(t, 2, currentTurn.ActionsRemaining(), "P1 should have 2 actions")
}

// TestFinalPhase_LastPlayerGetsUnlimitedActions verifies that when only
// one player remains active, they get unlimited actions.
func TestFinalPhase_LastPlayerGetsUnlimitedActions(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// Only P1 has plants
	p1.Resources().Set(shared.Resources{Plants: 24, Credits: 10})
	p2.Resources().Set(shared.Resources{Plants: 0, Credits: 10})

	maxAllGlobalParams(t, g)

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")
	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm")

	testutil.AssertEqual(t, shared.GamePhaseFinalPhase, g.CurrentPhase(), "Should be in final phase")

	// P1 is the only active player → should have unlimited actions (-1)
	currentTurn := g.CurrentTurn()
	testutil.AssertEqual(t, p1ID, currentTurn.PlayerID(), "P1 should have current turn")
	testutil.AssertEqual(t, -1, currentTurn.ActionsRemaining(), "P1 should have unlimited actions as last player")
}
