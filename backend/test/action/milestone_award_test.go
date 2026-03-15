package action_test

import (
	"context"
	"testing"

	awardAction "terraforming-mars-backend/internal/action/award"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Claim Milestone ---

func TestClaimMilestone_Terraformer_LogsCorrectName(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Resources().SetTerraformRating(35)

	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "terraformer")
	testutil.AssertNoError(t, err, "Claiming terraformer milestone should succeed")

	diffs, err := stateRepo.GetDiff(ctx, testGame.ID())
	testutil.AssertNoError(t, err, "GetDiff should succeed")
	testutil.AssertTrue(t, len(diffs) > 0, "Should have at least one diff")

	lastDiff := diffs[len(diffs)-1]
	testutil.AssertEqual(t, "Terraformer", lastDiff.Source, "Source should use sentence-case name")
	testutil.AssertEqual(t, "Claimed Terraformer milestone", lastDiff.Description, "Description should use sentence-case name")
	testutil.AssertEqual(t, string(shared.SourceTypeMilestone), string(lastDiff.SourceType), "SourceType should be milestone")
}

func TestClaimMilestone_Gardener_LogsCorrectName(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Place 3 greenery tiles to meet Gardener requirement
	positions := []shared.HexPosition{
		{Q: 0, R: 0, S: 0},
		{Q: 1, R: -1, S: 0},
		{Q: 2, R: -2, S: 0},
	}
	for _, pos := range positions {
		err := testGame.Board().UpdateTileOccupancy(ctx, pos, board.TileOccupant{
			Type: shared.ResourceGreeneryTile,
		}, playerID)
		if err != nil {
			t.Fatalf("Failed to place greenery: %v", err)
		}
	}

	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "gardener")
	testutil.AssertNoError(t, err, "Claiming gardener milestone should succeed")

	diffs, err := stateRepo.GetDiff(ctx, testGame.ID())
	testutil.AssertNoError(t, err, "GetDiff should succeed")
	testutil.AssertTrue(t, len(diffs) > 0, "Should have at least one diff")

	lastDiff := diffs[len(diffs)-1]
	testutil.AssertEqual(t, "Gardener", lastDiff.Source, "Source should use sentence-case name")
	testutil.AssertEqual(t, "Claimed Gardener milestone", lastDiff.Description, "Description should use sentence-case name")
}

func TestClaimMilestone_InvalidType(t *testing.T) {
	_, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, "some-game", playerID, "nonexistent")
	testutil.AssertError(t, err, "Invalid milestone type should fail")
}

func TestClaimMilestone_InsufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 0)
	p.Resources().SetTerraformRating(35)

	action := milestoneAction.NewClaimMilestoneAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "terraformer")
	testutil.AssertError(t, err, "Claiming milestone with 0 credits should fail")
}

// --- Fund Award ---

func TestFundAward_Landlord_LogsCorrectName(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	action := awardAction.NewFundAwardAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "landlord")
	testutil.AssertNoError(t, err, "Funding landlord award should succeed")

	diffs, err := stateRepo.GetDiff(ctx, testGame.ID())
	testutil.AssertNoError(t, err, "GetDiff should succeed")
	testutil.AssertTrue(t, len(diffs) > 0, "Should have at least one diff")

	lastDiff := diffs[len(diffs)-1]
	testutil.AssertEqual(t, "Landlord", lastDiff.Source, "Source should use sentence-case name")
	testutil.AssertEqual(t, "Funded Landlord award", lastDiff.Description, "Description should use sentence-case name")
	testutil.AssertEqual(t, string(shared.SourceTypeAward), string(lastDiff.SourceType), "SourceType should be award")
}

func TestFundAward_Scientist_LogsCorrectName(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	action := awardAction.NewFundAwardAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "scientist")
	testutil.AssertNoError(t, err, "Funding scientist award should succeed")

	diffs, err := stateRepo.GetDiff(ctx, testGame.ID())
	testutil.AssertNoError(t, err, "GetDiff should succeed")
	testutil.AssertTrue(t, len(diffs) > 0, "Should have at least one diff")

	lastDiff := diffs[len(diffs)-1]
	testutil.AssertEqual(t, "Scientist", lastDiff.Source, "Source should use sentence-case name")
	testutil.AssertEqual(t, "Funded Scientist award", lastDiff.Description, "Description should use sentence-case name")
}

func TestFundAward_InvalidType(t *testing.T) {
	_, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	action := awardAction.NewFundAwardAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, "some-game", playerID, "nonexistent")
	testutil.AssertError(t, err, "Invalid award type should fail")
}

func TestFundAward_InsufficientCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 0)

	action := awardAction.NewFundAwardAction(repo, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "landlord")
	testutil.AssertError(t, err, "Funding award with 0 credits should fail")
}

// --- Confirm Award Fund (Free) ---

func TestConfirmAwardFund_Success(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 10)

	// Set up pending award fund selection
	p.Selection().SetPendingAwardFundSelection(&shared.PendingAwardFundSelection{
		AvailableAwards: []string{"landlord", "banker", "scientist"},
		Source:          "corporation-starting-action",
	})

	action := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "banker")
	testutil.AssertNoError(t, err, "ConfirmAwardFund should succeed")

	// Award should be funded
	testutil.AssertTrue(t, testGame.Awards().IsFunded(shared.AwardBanker), "Banker award should be funded")

	// Credits should be unchanged (free)
	testutil.AssertEqual(t, 10, p.Resources().Get().Credits, "Credits should not be deducted")

	// Pending selection should be cleared
	testutil.AssertTrue(t, p.Selection().GetPendingAwardFundSelection() == nil, "Pending selection should be cleared")
}

func TestConfirmAwardFund_InvalidAwardType(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	p.Selection().SetPendingAwardFundSelection(&shared.PendingAwardFundSelection{
		AvailableAwards: []string{"landlord"},
		Source:          "corporation-starting-action",
	})

	action := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "nonexistent")
	testutil.AssertError(t, err, "Invalid award type should fail")
}

func TestConfirmAwardFund_AwardNotInAvailableList(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	p.Selection().SetPendingAwardFundSelection(&shared.PendingAwardFundSelection{
		AvailableAwards: []string{"landlord"},
		Source:          "corporation-starting-action",
	})

	action := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "banker")
	testutil.AssertError(t, err, "Award not in available list should fail")
}

func TestConfirmAwardFund_NoPendingSelection(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	action := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "landlord")
	testutil.AssertError(t, err, "Should fail without pending selection")
}
