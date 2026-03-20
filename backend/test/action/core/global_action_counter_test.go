package core_test

import (
	"context"
	"testing"

	"path/filepath"
	"runtime"

	cardAction "terraforming-mars-backend/internal/action/card"
	gameaction "terraforming-mars-backend/internal/action/game"
	spAction "terraforming-mars-backend/internal/action/standard_project"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/standardprojects"
	"terraforming-mars-backend/test/testutil"
)

func TestGlobalActionCounterStartsAtZero(t *testing.T) {
	testGame, _, _, _, _ := testutil.SetupTwoPlayerGame(t)

	turn := testGame.CurrentTurn()
	testutil.AssertEqual(t, 0, turn.GlobalActionCounter(), "Global action counter should start at 0")
}

func TestPlayCardIncrementsGlobalActionCounter(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 100)
	p.Hand().AddCard(testutil.CardID("Power Plant"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}

	err := playAction.Execute(context.Background(), testGame.ID(), playerID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing card should succeed")

	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Global action counter should be 1 after playing a card")
}

func createStdProjRegistry(t *testing.T) standardprojects.StandardProjectRegistry {
	t.Helper()
	_, currentFile, _, _ := runtime.Caller(0)
	stdProjPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "assets", "terraforming_mars_standard_projects.json")
	stdProjData, err := standardprojects.LoadStandardProjectsFromJSON(stdProjPath)
	if err != nil {
		t.Fatalf("Failed to load standard projects: %v", err)
	}
	return standardprojects.NewInMemoryStandardProjectRegistry(stdProjData)
}

func TestStandardProjectIncrementsGlobalActionCounter(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), p, 100)

	stdProjRegistry := createStdProjRegistry(t)
	buildAction := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)

	err := buildAction.Execute(context.Background(), testGame.ID(), playerID, "power-plant")
	testutil.AssertNoError(t, err, "Building power plant should succeed")

	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Global action counter should be 1 after standard project")
}

func TestSkipIncrementsGlobalActionCounter(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	err := testGame.SetCurrentTurn(context.Background(), player1ID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	err = skipAction.Execute(context.Background(), testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "SKIP should succeed")

	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Global action counter should be 1 after skip")
}

func TestPassIncrementsGlobalActionCounter(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	// Player 1 passes (has 2 actions = pass)
	err := skipAction.Execute(context.Background(), testGame.ID(), player1ID)
	testutil.AssertNoError(t, err, "PASS should succeed")

	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Global action counter should be 1 after pass")
}

func TestMultipleActionsIncrementCounterSequentially(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p1, _ := testGame.GetPlayer(player1ID)
	testutil.SetPlayerCredits(context.Background(), p1, 200)
	p1.Hand().AddCard(testutil.CardID("Power Plant"))
	p1.Hand().AddCard(testutil.CardID("Asteroid Mining"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)

	// Player 1 plays first card (counter: 0 -> 1)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Power Plant"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing first card should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Counter should be 1 after first action")

	// Player 1 plays second card (counter: 1 -> 2), auto-advances to player 2
	payment = cardAction.PaymentRequest{Credits: 30}
	err = playAction.Execute(context.Background(), testGame.ID(), player1ID, testutil.CardID("Asteroid Mining"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing second card should succeed")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().GlobalActionCounter(), "Counter should be 2 after second action")

	// Player 2 uses standard project (counter: 2 -> 3)
	p2, _ := testGame.GetPlayer(player2ID)
	testutil.SetPlayerCredits(context.Background(), p2, 200)

	stdProjRegistry := createStdProjRegistry(t)
	buildAction := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err = buildAction.Execute(context.Background(), testGame.ID(), player2ID, "power-plant")
	testutil.AssertNoError(t, err, "Player 2 building power plant should succeed")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().GlobalActionCounter(), "Counter should be 3 after player 2 action")
}

func TestCounterIncrementsDuringUnlimitedActions(t *testing.T) {
	testGame, repo, cardRegistry, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	// Player 2 passes so player 1 gets unlimited actions
	p2, _ := testGame.GetPlayer(player2ID)
	p2.SetPassed(true)

	// Grant unlimited actions to player 1
	err := testGame.SetCurrentTurn(context.Background(), player1ID, -1)
	testutil.AssertNoError(t, err, "Setting unlimited actions should succeed")

	p1, _ := testGame.GetPlayer(player1ID)
	testutil.SetPlayerCredits(context.Background(), p1, 200)

	// Play standard project with unlimited actions (counter should still increment)
	stdProjRegistry := createStdProjRegistry(t)
	buildAction := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)

	err = buildAction.Execute(context.Background(), testGame.ID(), player1ID, "power-plant")
	testutil.AssertNoError(t, err, "First action with unlimited should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().GlobalActionCounter(), "Counter should increment even with unlimited actions")

	err = buildAction.Execute(context.Background(), testGame.ID(), player1ID, "power-plant")
	testutil.AssertNoError(t, err, "Second action with unlimited should succeed")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().GlobalActionCounter(), "Counter should be 2 after second unlimited action")
}
