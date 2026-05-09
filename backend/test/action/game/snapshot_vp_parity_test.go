package game_test

import (
	"context"
	"testing"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestSnapshotEnricher_ParityWithFinalScoring verifies that the VP breakdown stored
// on each history snapshot (via the enricher) matches what FinalScoringAction would
// compute on the same game state. This is the core single-source-of-truth guarantee:
// any future change to scoring logic flows to both code paths automatically.
func TestSnapshotEnricher_ParityWithFinalScoring(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := context.Background()
	awardRegistry := testutil.CreateTestAwardRegistry()
	milestoneRegistry := testutil.CreateTestMilestoneRegistry()

	repo.DataStore().SetSnapshotEnricher(func(state *datastore.GameState) map[string]shared.VPBreakdown {
		live, err := repo.Get(ctx, state.ID)
		if err != nil || live == nil {
			return nil
		}
		return gameaction.ComputePlayerVPBreakdowns(live, cardRegistry, awardRegistry, milestoneRegistry)
	})

	p1, _ := g.GetPlayer(playerIDs[0])
	p2, _ := g.GetPlayer(playerIDs[1])

	p1.Resources().Set(shared.Resources{Heat: 30, Credits: 100})
	p2.Resources().Set(shared.Resources{Heat: 5, Credits: 100})

	if err := g.Awards().FundAward(ctx, "thermalist", playerIDs[0], 8); err != nil {
		t.Fatalf("Fund Thermalist award: %v", err)
	}

	if err := g.SetGeneration(ctx, g.Generation()+1); err != nil {
		t.Fatalf("Bump generation to trigger snapshot: %v", err)
	}

	expected := gameaction.ComputePlayerVPBreakdowns(g, cardRegistry, awardRegistry, milestoneRegistry)
	testutil.AssertTrue(t, expected != nil, "ComputePlayerVPBreakdowns should produce a result")

	history, err := repo.DataStore().GetGameHistory(g.ID())
	testutil.AssertNoError(t, err, "GetGameHistory")
	testutil.AssertTrue(t, len(history) > 0, "History should have at least one entry")

	latest := history[len(history)-1]
	testutil.AssertTrue(t, latest.VPBreakdowns != nil, "Latest snapshot should have VPBreakdowns populated by the enricher")

	for _, pid := range playerIDs {
		exp := expected[pid]
		got, ok := latest.VPBreakdowns[pid]
		testutil.AssertTrue(t, ok, "VPBreakdowns should contain entry for player "+pid)
		testutil.AssertEqual(t, exp.TotalVP, got.TotalVP, "TotalVP parity for player "+pid)
		testutil.AssertEqual(t, exp.AwardVP, got.AwardVP, "AwardVP parity for player "+pid)
		testutil.AssertEqual(t, exp.TerraformRating, got.TerraformRating, "TerraformRating parity for player "+pid)
	}

	winnerVP := expected[playerIDs[0]].AwardVP
	loserVP := expected[playerIDs[1]].AwardVP
	testutil.AssertTrue(t, winnerVP > loserVP, "Player with more heat should outscore on Thermalist")
}
