package game_test

import (
	"context"
	"testing"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestFinalScoring_CardVPIncluded verifies that FinalScoringAction produces
// non-zero CardVP when players have VP-granting cards played.
func TestFinalScoring_CardVPIncluded(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	// Player 1: play Colonizer Training Camp (001) = 2 VP fixed
	ctcID := testutil.CardID("Colonizer Training Camp")
	p1.PlayedCards().AddCard(ctcID, "Colonizer Training Camp", "automated", []string{"building", "jovian"})

	// Player 1: play Predators (024) with 4 animals = 4 VP
	predatorsID := testutil.CardID("Predators")
	p1.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p1.Resources().AddToStorage(predatorsID, 4)

	// Player 2: play Asteroid Mining Consortium (002) = 1 VP fixed
	amcID := testutil.CardID("Asteroid Mining Consortium")
	p2.PlayedCards().AddCard(amcID, "Asteroid Mining Consortium", "automated", []string{"jovian"})

	// Max global parameters so game can end
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

	// Execute final scoring
	action := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	err := action.Execute(ctx, g.ID())
	if err != nil {
		t.Fatalf("FinalScoringAction failed: %v", err)
	}

	// Check final scores
	scores := g.GetFinalScores()
	if scores == nil {
		t.Fatal("Final scores are nil after FinalScoringAction")
	}

	if len(scores) != 2 {
		t.Fatalf("Expected 2 scores, got %d", len(scores))
	}

	// Find player 1's score
	var p1Score, p2Score *struct {
		cardVP  int
		totalVP int
	}
	for _, s := range scores {
		if s.PlayerID == p1ID {
			p1Score = &struct {
				cardVP  int
				totalVP int
			}{s.Breakdown.CardVP, s.Breakdown.TotalVP}
		}
		if s.PlayerID == p2ID {
			p2Score = &struct {
				cardVP  int
				totalVP int
			}{s.Breakdown.CardVP, s.Breakdown.TotalVP}
		}
	}

	if p1Score == nil {
		t.Fatal("Player 1 score not found in final scores")
	}
	if p2Score == nil {
		t.Fatal("Player 2 score not found in final scores")
	}

	// Player 1: Colonizer Training Camp (2) + Predators w/ 4 animals (4) = 6 card VP
	if p1Score.cardVP != 6 {
		t.Errorf("Player 1 CardVP: expected 6, got %d", p1Score.cardVP)
	}

	// Player 2: Asteroid Mining Consortium (1) = 1 card VP
	if p2Score.cardVP != 1 {
		t.Errorf("Player 2 CardVP: expected 1, got %d", p2Score.cardVP)
	}

	// Total VP should include CardVP
	if p1Score.totalVP < p1Score.cardVP {
		t.Errorf("Player 1 TotalVP (%d) should be >= CardVP (%d)", p1Score.totalVP, p1Score.cardVP)
	}
}

// TestFinalScoring_GreeneryAndCityVP verifies tile VP in final scoring.
func TestFinalScoring_GreeneryAndCityVP(t *testing.T) {
	g, repo, cardRegistry, p1ID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	// Place greenery and city tiles for player 1
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceGreeneryTile,
	}, p1ID)
	if err != nil {
		t.Fatalf("Failed to place greenery: %v", err)
	}

	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 1, R: -1, S: 0}, board.TileOccupant{
		Type: shared.ResourceCityTile,
	}, p1ID)
	if err != nil {
		t.Fatalf("Failed to place city: %v", err)
	}

	// Max global parameters
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

	action := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	err = action.Execute(ctx, g.ID())
	if err != nil {
		t.Fatalf("FinalScoringAction failed: %v", err)
	}

	scores := g.GetFinalScores()
	for _, s := range scores {
		if s.PlayerID == p1ID {
			if s.Breakdown.GreeneryVP != 1 {
				t.Errorf("GreeneryVP: expected 1, got %d", s.Breakdown.GreeneryVP)
			}
			// City at (1,-1,0) adjacent to greenery at (0,0,0) = 1 city VP
			if s.Breakdown.CityVP != 1 {
				t.Errorf("CityVP: expected 1, got %d", s.Breakdown.CityVP)
			}
			return
		}
	}
	t.Fatal("Player 1 score not found")
}
