package core_test

import (
	"context"
	"testing"

	awardAction "terraforming-mars-backend/internal/action/award"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdAction "terraforming-mars-backend/internal/action/standard_project"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
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
	err := action.Execute(context.Background(), testGame.ID(), playerID)

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
	err := action.Execute(context.Background(), testGame.ID(), playerID)

	testutil.AssertError(t, err, "Convert plants should be rejected during production phase")
}

func TestSellPatents_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	action := stdAction.NewSellPatentsAction(repo, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID)

	testutil.AssertError(t, err, "Sell patents should be rejected during production phase")
}

func TestBuildPowerPlant_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), player, 20)

	action := stdAction.NewBuildPowerPlantAction(repo, cardRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID)

	testutil.AssertError(t, err, "Build power plant should be rejected during production phase")
}

func TestBuildAquifer_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(context.Background(), player, 20)

	action := stdAction.NewBuildAquiferAction(repo, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID)

	testutil.AssertError(t, err, "Build aquifer should be rejected during production phase")
}

func TestClaimMilestone_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "terraformer")

	testutil.AssertError(t, err, "Claim milestone should be rejected during production phase")
}

func TestFundAward_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	action := awardAction.NewFundAwardAction(repo, cardRegistry, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, "landlord")

	testutil.AssertError(t, err, "Fund award should be rejected during production phase")
}

func TestConfirmSellPatents_RejectsDuringProductionPhase(t *testing.T) {
	testGame, repo, playerID := setupProductionPhaseGame(t)
	logger := testutil.TestLogger()

	action := confirmAction.NewConfirmSellPatentsAction(repo, nil, logger)
	err := action.Execute(context.Background(), testGame.ID(), playerID, []string{})

	testutil.AssertError(t, err, "Confirm sell patents should be rejected during production phase")
}
