package core_test

import (
	"context"
	"testing"

	"path/filepath"
	"runtime"

	awardAction "openmars/internal/action/award"
	confirmAction "openmars/internal/action/confirmation"
	milestoneAction "openmars/internal/action/milestone"
	resconvAction "openmars/internal/action/resource_conversion"
	stdAction "openmars/internal/action/standard_project"
	"openmars/internal/game"
	"openmars/internal/game/shared"
	"openmars/internal/game/standardproject"
	"openmars/test/testutil"
)

func setupProductionPhaseGame(t *testing.T) (*game.Game, game.GameRepository, string) {
	t.Helper()

	testGame, repo, cardRegistry, playerID := setupActiveGame(t)
	_ = cardRegistry

	ctx := context.Background()
	err := testGame.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw)
	if err != nil {
		t.Fatalf("Failed to set production phase: %v", err)
	}

	return testGame, repo, playerID
}

func TestConvertHeat_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(context.Background(), player, 8)

	action := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, shared.NativePayment(shared.ResourceHeat, 8))

	testutil.AssertError(t, err, "Convert heat should be rejected during production phase")
}

func TestConvertPlantsToGreenery_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	player, _ := testGame.GetPlayer(playerID)
	resources := player.Resources().Get()
	resources.Plants = 8
	player.Resources().Set(resources)

	action := resconvAction.NewConvertPlantsToGreeneryAction(repo, cardRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, shared.NativePayment(shared.ResourcePlant, 8))

	testutil.AssertError(t, err, "Convert plants should be rejected during production phase")
}

func createPhaseTestStdProjRegistry(t *testing.T) standardproject.StandardProjectRegistry {
	t.Helper()
	_, currentFile, _, _ := runtime.Caller(0)
	stdProjPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "assets", "standard_projects.json")
	stdProjData, err := standardproject.LoadStandardProjectsFromJSON(stdProjPath)
	if err != nil {
		t.Fatalf("Failed to load standard projects: %v", err)
	}
	return standardproject.NewInMemoryStandardProjectRegistry(stdProjData)
}

func TestSellPatents_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	stdProjRegistry := createPhaseTestStdProjRegistry(t)
	action := stdAction.NewExecuteStandardProjectAction(repo, nil, stdProjRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "sell-patents", shared.NativePayment(shared.ResourceCredit, shared.StandardProjectCost[shared.StandardProject("sell-patents")]))

	testutil.AssertError(t, err, "Sell patents should be rejected during production phase")
}

func TestBuildPowerPlant_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), player, 20)

	stdProjRegistry := createPhaseTestStdProjRegistry(t)
	action := stdAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "power-plant", shared.NativePayment(shared.ResourceCredit, shared.StandardProjectCost[shared.StandardProject("power-plant")]))

	testutil.AssertError(t, err, "Build power plant should be rejected during production phase")
}

func TestBuildAquifer_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), player, 20)

	stdProjRegistry := createPhaseTestStdProjRegistry(t)
	action := stdAction.NewExecuteStandardProjectAction(repo, nil, stdProjRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "aquifer", shared.NativePayment(shared.ResourceCredit, shared.StandardProjectCost[shared.StandardProject("aquifer")]))

	testutil.AssertError(t, err, "Build aquifer should be rejected during production phase")
}

func TestClaimMilestone_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	milestoneRegistry := testutil.CreateTestMilestoneRegistry()
	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, nil, milestoneRegistry, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "terraformer", shared.NativePayment(shared.ResourceCredit, 8))

	testutil.AssertError(t, err, "Claim milestone should be rejected during production phase")
}

func TestFundAward_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	awardRegistry := testutil.CreateTestAwardRegistry()
	action := awardAction.NewFundAwardAction(repo, cardRegistry, nil, awardRegistry, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "landlord", shared.NativePayment(shared.ResourceCredit, 8))

	testutil.AssertError(t, err, "Fund award should be rejected during production phase")
}

func TestConfirmSellPatents_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	action := confirmAction.NewConfirmSellPatentsAction(repo, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, []string{})

	testutil.AssertError(t, err, "Confirm sell patents should be rejected during production phase")
}
