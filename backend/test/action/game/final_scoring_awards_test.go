package game_test

import (
	"context"
	"testing"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// maxAllGlobalParams sets all global parameters to their maximum values.
func maxAllGlobalParams(t *testing.T, g *game.Game) {
	t.Helper()
	ctx := context.Background()
	gp := g.GlobalParameters()
	for !gp.IsMaxed() {
		if gp.Temperature() < 8 {
			if _, err := gp.IncreaseTemperature(ctx, 1, ""); err != nil {
				t.Fatalf("Failed to increase temperature: %v", err)
			}
		}
		if gp.Oxygen() < 14 {
			if _, err := gp.IncreaseOxygen(ctx, 1, ""); err != nil {
				t.Fatalf("Failed to increase oxygen: %v", err)
			}
		}
		if gp.Oceans() < 9 {
			if _, err := gp.PlaceOcean(ctx, ""); err != nil {
				t.Fatalf("Failed to place ocean: %v", err)
			}
		}
	}
}

// TestFinalScoring_ThermalistAwardReflectsFinalProduction verifies that the Thermalist award
// (most heat resources) is correctly scored based on post-production resources.
// Player A has more heat initially, but Player B has much higher heat production,
// so after final production Player B should win the Thermalist award.
func TestFinalScoring_ThermalistAwardReflectsFinalProduction(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := context.Background()
	logger := testutil.TestLogger()
	awardRegistry := testutil.CreateTestAwardRegistry()
	milestoneRegistry := testutil.CreateTestMilestoneRegistry()

	p1, _ := g.GetPlayer(playerIDs[0])
	p2, _ := g.GetPlayer(playerIDs[1])

	// Player A: 20 heat, 0 heat production
	p1.Resources().Set(shared.Resources{Heat: 20, Credits: 100})
	p1.Resources().SetProduction(shared.Production{Heat: 0, Credits: 0})

	// Player B: 5 heat, 30 heat production
	p2.Resources().Set(shared.Resources{Heat: 5, Credits: 100})
	p2.Resources().SetProduction(shared.Production{Heat: 30, Credits: 0})

	// Fund the Thermalist award
	err := g.Awards().FundAward(ctx, "thermalist", playerIDs[0], 8)
	testutil.AssertNoError(t, err, "Fund Thermalist award")

	// Before production: P1 has 20 heat, P2 has 5 heat → P1 leads
	testutil.AssertTrue(t, p1.Resources().Get().Heat > p2.Resources().Get().Heat,
		"Before production, Player A should have more heat than Player B")

	// Simulate final production (mimicking ExecuteFinalProductionPhase)
	for _, pID := range playerIDs {
		p, _ := g.GetPlayer(pID)
		r := p.Resources().Get()
		prod := p.Resources().Production()
		tr := p.Resources().TerraformRating()
		p.Resources().Set(shared.Resources{
			Credits:  r.Credits + prod.Credits + tr,
			Steel:    r.Steel + prod.Steel,
			Titanium: r.Titanium + prod.Titanium,
			Plants:   r.Plants + prod.Plants,
			Energy:   prod.Energy,
			Heat:     r.Heat + r.Energy + prod.Heat,
		})
	}

	// After production: P1 has 20 heat, P2 has 35 heat → P2 leads
	testutil.AssertTrue(t, p2.Resources().Get().Heat > p1.Resources().Get().Heat,
		"After production, Player B should have more heat than Player A")

	maxAllGlobalParams(t, g)

	action := gameaction.NewFinalScoringAction(repo, cardRegistry, awardRegistry, milestoneRegistry, logger)
	err = action.Execute(ctx, g.ID())
	testutil.AssertNoError(t, err, "FinalScoringAction")

	scores := g.GetFinalScores()
	testutil.AssertTrue(t, scores != nil, "Final scores should not be nil")

	var p1AwardVP, p2AwardVP int
	for _, s := range scores {
		if s.PlayerID == playerIDs[0] {
			p1AwardVP = s.Breakdown.AwardVP
		}
		if s.PlayerID == playerIDs[1] {
			p2AwardVP = s.Breakdown.AwardVP
		}
	}

	// Player B should get 1st place (5 VP) for Thermalist, Player A gets 2nd place (2 VP)
	testutil.AssertEqual(t, 5, p2AwardVP, "Player B should get 5 VP for 1st place Thermalist")
	testutil.AssertEqual(t, 2, p1AwardVP, "Player A should get 2 VP for 2nd place Thermalist")
}

// TestFinalScoring_BankerAwardUsesProduction verifies that the Banker award
// is scored based on MC production (not MC resources).
func TestFinalScoring_BankerAwardUsesProduction(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := context.Background()
	logger := testutil.TestLogger()
	awardRegistry := testutil.CreateTestAwardRegistry()
	milestoneRegistry := testutil.CreateTestMilestoneRegistry()

	p1, _ := g.GetPlayer(playerIDs[0])
	p2, _ := g.GetPlayer(playerIDs[1])

	// Player A: high credits but low MC production
	p1.Resources().Set(shared.Resources{Credits: 200})
	p1.Resources().SetProduction(shared.Production{Credits: 2})

	// Player B: low credits but high MC production
	p2.Resources().Set(shared.Resources{Credits: 10})
	p2.Resources().SetProduction(shared.Production{Credits: 15})

	// Fund the Banker award
	err := g.Awards().FundAward(ctx, "banker", playerIDs[0], 8)
	testutil.AssertNoError(t, err, "Fund Banker award")

	maxAllGlobalParams(t, g)

	action := gameaction.NewFinalScoringAction(repo, cardRegistry, awardRegistry, milestoneRegistry, logger)
	err = action.Execute(ctx, g.ID())
	testutil.AssertNoError(t, err, "FinalScoringAction")

	scores := g.GetFinalScores()
	testutil.AssertTrue(t, scores != nil, "Final scores should not be nil")

	var p1AwardVP, p2AwardVP int
	for _, s := range scores {
		if s.PlayerID == playerIDs[0] {
			p1AwardVP = s.Breakdown.AwardVP
		}
		if s.PlayerID == playerIDs[1] {
			p2AwardVP = s.Breakdown.AwardVP
		}
	}

	// Banker checks MC production, not credits
	// Player B should win (15 production vs 2 production)
	testutil.AssertEqual(t, 5, p2AwardVP, "Player B should get 5 VP for 1st place Banker")
	testutil.AssertEqual(t, 2, p1AwardVP, "Player A should get 2 VP for 2nd place Banker")
}

// TestFinalScoring_MinerAwardReflectsFinalResources verifies that the Miner award
// (most steel + titanium resources) scores correctly when production changes the leader.
func TestFinalScoring_MinerAwardReflectsFinalResources(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := context.Background()
	logger := testutil.TestLogger()
	awardRegistry := testutil.CreateTestAwardRegistry()
	milestoneRegistry := testutil.CreateTestMilestoneRegistry()

	p1, _ := g.GetPlayer(playerIDs[0])
	p2, _ := g.GetPlayer(playerIDs[1])

	// Player A: 10 steel + 5 titanium = 15 total, no production
	p1.Resources().Set(shared.Resources{Steel: 10, Titanium: 5, Credits: 100})
	p1.Resources().SetProduction(shared.Production{Steel: 0, Titanium: 0})

	// Player B: 3 steel + 2 titanium = 5 total, high production
	p2.Resources().Set(shared.Resources{Steel: 3, Titanium: 2, Credits: 100})
	p2.Resources().SetProduction(shared.Production{Steel: 8, Titanium: 5})

	// Fund the Miner award
	err := g.Awards().FundAward(ctx, "miner", playerIDs[0], 8)
	testutil.AssertNoError(t, err, "Fund Miner award")

	// Before production: P1=15, P2=5
	testutil.AssertTrue(t, p1.Resources().Get().Steel+p1.Resources().Get().Titanium >
		p2.Resources().Get().Steel+p2.Resources().Get().Titanium,
		"Before production, Player A should have more steel+titanium")

	// Simulate final production
	for _, pID := range playerIDs {
		p, _ := g.GetPlayer(pID)
		r := p.Resources().Get()
		prod := p.Resources().Production()
		tr := p.Resources().TerraformRating()
		p.Resources().Set(shared.Resources{
			Credits:  r.Credits + prod.Credits + tr,
			Steel:    r.Steel + prod.Steel,
			Titanium: r.Titanium + prod.Titanium,
			Plants:   r.Plants + prod.Plants,
			Energy:   prod.Energy,
			Heat:     r.Heat + r.Energy + prod.Heat,
		})
	}

	// After production: P1=15, P2=18
	testutil.AssertTrue(t, p2.Resources().Get().Steel+p2.Resources().Get().Titanium >
		p1.Resources().Get().Steel+p1.Resources().Get().Titanium,
		"After production, Player B should have more steel+titanium")

	maxAllGlobalParams(t, g)

	action := gameaction.NewFinalScoringAction(repo, cardRegistry, awardRegistry, milestoneRegistry, logger)
	err = action.Execute(ctx, g.ID())
	testutil.AssertNoError(t, err, "FinalScoringAction")

	scores := g.GetFinalScores()
	testutil.AssertTrue(t, scores != nil, "Final scores should not be nil")

	var p1AwardVP, p2AwardVP int
	for _, s := range scores {
		if s.PlayerID == playerIDs[0] {
			p1AwardVP = s.Breakdown.AwardVP
		}
		if s.PlayerID == playerIDs[1] {
			p2AwardVP = s.Breakdown.AwardVP
		}
	}

	testutil.AssertEqual(t, 5, p2AwardVP, "Player B should get 5 VP for 1st place Miner")
	testutil.AssertEqual(t, 2, p1AwardVP, "Player A should get 2 VP for 2nd place Miner")
}
