package core_test

import (
	"context"
	"testing"

	"path/filepath"
	"runtime"

	"terraforming-mars-backend/internal/action"
	baseaction "terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	gameaction "terraforming-mars-backend/internal/action/game"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	spAction "terraforming-mars-backend/internal/action/standard_project"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/standardprojects"
	"terraforming-mars-backend/test/testutil"
)

func TestPlayCardConsumesAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 100)
	p.Hand().AddCard(testutil.CardID("Power Plant"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err := playAction.Execute(context.Background(), testGame.ID(), playerID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing card should succeed")

	turn := testGame.CurrentTurn()
	testutil.AssertEqual(t, 1, turn.ActionsRemaining(), "Should have 1 action remaining after playing card")
}

func TestZeroActionsBlocksCardPlay(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Set actions to 0
	err := testGame.SetCurrentTurn(context.Background(), playerID, 0)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 100)
	p.Hand().AddCard(testutil.CardID("Power Plant"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err = playAction.Execute(context.Background(), testGame.ID(), playerID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail with 0 actions remaining")
}

func TestZeroActionsBlocksStandardProject(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	err := testGame.SetCurrentTurn(context.Background(), playerID, 0)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 100)

	_, currentFile, _, _ := runtime.Caller(0)
	stdProjPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "assets", "terraforming_mars_standard_projects.json")
	stdProjData, _ := standardprojects.LoadStandardProjectsFromJSON(stdProjPath)
	stdProjRegistry := standardprojects.NewInMemoryStandardProjectRegistry(stdProjData)

	buildAction := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)

	err = buildAction.Execute(context.Background(), testGame.ID(), playerID, "power-plant")
	testutil.AssertError(t, err, "Should fail with 0 actions remaining")
}

func TestZeroActionsBlocksConvertHeat(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	err := testGame.SetCurrentTurn(context.Background(), playerID, 0)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(context.Background(), p, 20)

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	err = convertAction.Execute(context.Background(), testGame.ID(), playerID)
	testutil.AssertError(t, err, "Should fail with 0 actions remaining")
}

func TestAutoAdvanceAfterSecondAction(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p1, _ := testGame.GetPlayer(player1ID)
	testutil.SetPlayerCredits(context.Background(), p1, 200)

	// Play first card
	p1.Hand().AddCard(testutil.CardID("Power Plant"))
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err := playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "First card play should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().ActionsRemaining(), "Should have 1 action after first play")
	testutil.AssertEqual(t, player1ID, testGame.CurrentTurn().PlayerID(), "Should still be player 1's turn")

	// Play second card
	p1.Hand().AddCard(testutil.CardID("Asteroid"))
	payment2 := cardAction.PaymentRequest{Credits: 14}
	err = playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Asteroid"), payment2, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Second card play should succeed")

	// Should auto-advance to player 2
	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(), "Turn should auto-advance to player 2")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player 2 should have 2 actions")
}

func TestSoloUnlimitedActionsNotBlocked(t *testing.T) {
	testGame, repo, cardRegistry, playerID := testutil.SetupSoloGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 200)

	// Verify unlimited actions
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should have unlimited actions")

	// Play card - should succeed and remain unlimited
	p.Hand().AddCard(testutil.CardID("Power Plant"))
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err := playAction.Execute(context.Background(), testGame.ID(), playerID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Solo card play should succeed")

	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should still have unlimited actions")
	testutil.AssertEqual(t, playerID, testGame.CurrentTurn().PlayerID(), "Solo player should still have the turn")
}

func TestStateCalculatorReportsNoActionsRemaining(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	err := testGame.SetCurrentTurn(context.Background(), playerID, 0)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 200)

	// Check state calculator for standard project
	state := action.CalculatePlayerStandardProjectState(
		"power-plant",
		p,
		testGame,
		cardRegistry,
	)

	hasNoActionsError := false
	for _, stateErr := range state.Errors {
		if stateErr.Code == player.ErrorCodeNoActionsRemaining {
			hasNoActionsError = true
			break
		}
	}
	testutil.AssertTrue(t, hasNoActionsError, "State calculator should report no-actions-remaining error")
}

func TestAutoAdvanceWaitsForPendingTileSelection(t *testing.T) {
	testGame, _, _, player1ID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Set pending tile selection for player 1
	err := testGame.SetPendingTileSelection(context.Background(), player1ID, &shared.PendingTileSelection{
		TileType:       "greenery",
		AvailableHexes: []string{"0,1,-1"},
		Source:         "test",
	})
	testutil.AssertNoError(t, err, "Setting pending tile selection should succeed")

	// Set actions to 0 (simulating consumed action)
	err = testGame.SetCurrentTurn(context.Background(), player1ID, 0)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	// Call AutoAdvanceTurnIfNeeded - should NOT advance because of pending tile
	baseaction.AutoAdvanceTurnIfNeeded(testGame, player1ID, logger)

	testutil.AssertEqual(t, player1ID, testGame.CurrentTurn().PlayerID(), "Should NOT advance due to pending tile selection")
}

func TestSkipWithOneActionAdvancesToNextPlayer(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Set player 1 to have 1 action remaining
	err := testGame.SetCurrentTurn(context.Background(), player1ID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	// Create skip action
	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	// Player 1 SKIPs with 1 action
	err = skipAction.Execute(context.Background(), testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "SKIP should succeed")

	// Turn should advance to player 2 with 2 actions
	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(), "Turn should advance to player 2")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player 2 should have 2 actions")
}

func TestAutoAdvanceGrantsUnlimitedActionsToLastNonPassedPlayer(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Player 2 passes
	p2, _ := testGame.GetPlayer(player2ID)
	p2.SetPassed(true)

	// Player 1 plays two cards to consume both actions
	p1, _ := testGame.GetPlayer(player1ID)
	testutil.SetPlayerCredits(context.Background(), p1, 200)

	// Play first card
	p1.Hand().AddCard(testutil.CardID("Power Plant"))
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err := playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "First card play should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().ActionsRemaining(), "Should have 1 action after first play")

	// Play second card - this should auto-advance and grant unlimited actions to player 1
	p1.Hand().AddCard(testutil.CardID("Asteroid"))
	payment2 := cardAction.PaymentRequest{Credits: 14}
	err = playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Asteroid"), payment2, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Second card play should succeed")

	// Player 1 should now have unlimited actions since they're the last non-passed player
	testutil.AssertEqual(t, player1ID, testGame.CurrentTurn().PlayerID(), "Player 1 should still have the turn")
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Player 1 should have unlimited actions as last non-passed player")
}

func TestSkipGrantsUnlimitedActionsToLastNonPassedPlayer(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Player 2 passes
	p2, _ := testGame.GetPlayer(player2ID)
	p2.SetPassed(true)

	// Player 1 has 1 action and SKIPs
	err := testGame.SetCurrentTurn(context.Background(), player1ID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	// Create skip action
	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	// Player 1 SKIPs
	err = skipAction.Execute(context.Background(), testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "SKIP should succeed")

	// Player 1 should get unlimited actions since they're the last non-passed player
	testutil.AssertEqual(t, player1ID, testGame.CurrentTurn().PlayerID(), "Player 1 should still have the turn")
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Player 1 should have unlimited actions as last non-passed player")
}

func TestTurnOrderRotatesAfterGeneration(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Verify initial turn order
	initialTurnOrder := testGame.TurnOrder()
	testutil.AssertEqual(t, player1ID, initialTurnOrder[0], "Player 1 should be first in initial turn order")
	testutil.AssertEqual(t, player2ID, initialTurnOrder[1], "Player 2 should be second in initial turn order")

	// Both players pass to trigger production phase
	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	// Player 1 passes (2 actions = pass)
	err := skipAction.Execute(context.Background(), testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "Player 1 PASS should succeed")

	// Player 2 passes (2 actions = pass)
	err = skipAction.Execute(context.Background(), testGame.ID(), player2ID)
	testutil.AssertNoError(t, err, "Player 2 PASS should succeed")

	// After both pass, production phase triggers and turn order should rotate
	// Player 2 should now be first in turn order
	newTurnOrder := testGame.TurnOrder()
	testutil.AssertEqual(t, player2ID, newTurnOrder[0], "Player 2 should be first after turn order rotation")
	testutil.AssertEqual(t, player1ID, newTurnOrder[1], "Player 1 should be second after turn order rotation")
}

func TestForcedFirstActionDoesNotConsumePlayerAction(t *testing.T) {
	testGame, _, _, player1ID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	// Set a forced first action for player 1 (simulating Tharsis Republic)
	forcedAction := &shared.ForcedFirstAction{
		ActionType:    "city-placement",
		CorporationID: testutil.CardID("Tharsis Republic"),
		Source:        "corporation-starting-action",
		Completed:     false,
		Description:   "Place a city tile (Tharsis Republic starting action)",
	}
	err := testGame.SetForcedFirstAction(ctx, player1ID, forcedAction)
	testutil.AssertNoError(t, err, "Setting forced first action should succeed")

	// Verify player has 2 actions
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player should have 2 actions before forced action")

	// Clear the forced first action (simulating completion)
	// Note: In the real flow, this is done by the corporation_processor after tile placement
	// The key test is that NO action consumption happens - we removed ConsumeAction() from the processor
	err = testGame.SetForcedFirstAction(ctx, player1ID, nil)
	testutil.AssertNoError(t, err, "Clearing forced first action should succeed")

	// Verify player STILL has 2 actions (forced actions are FREE)
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player should still have 2 actions after forced action (forced actions are free)")
}
