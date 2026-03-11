package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestLandlordAwardExcludesOceans verifies that Landlord counts cities + greeneries only.
func TestLandlordAwardExcludesOceans(t *testing.T) {
	g, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	// Place 1 city owned by player
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceCityTile,
	}, playerID)
	if err != nil {
		t.Fatalf("Failed to place city: %v", err)
	}

	// Place 1 greenery owned by player
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 2, R: -2, S: 0}, board.TileOccupant{
		Type: shared.ResourceGreeneryTile,
	}, playerID)
	if err != nil {
		t.Fatalf("Failed to place greenery: %v", err)
	}

	// Place 2 ocean tiles (should NOT count for Landlord)
	oceanPositions := []shared.HexPosition{
		{Q: 3, R: -3, S: 0},
		{Q: 4, R: -4, S: 0},
	}
	for _, pos := range oceanPositions {
		err := g.Board().UpdateTileOccupancy(ctx, pos, board.TileOccupant{
			Type: shared.ResourceOceanTile,
		}, playerID)
		if err != nil {
			t.Fatalf("Failed to place ocean: %v", err)
		}
	}

	score := gamecards.CalculateAwardScore(shared.AwardLandlord, p, g.Board(), cardRegistry)

	// Should count 1 city + 1 greenery = 2, NOT 4 (which would include oceans)
	if score != 2 {
		t.Fatalf("expected Landlord score=2 (1 city + 1 greenery), got %d", score)
	}
}

// TestLandlordAwardZeroWithOnlyOceans verifies Landlord returns 0 when player only has oceans.
func TestLandlordAwardZeroWithOnlyOceans(t *testing.T) {
	g, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	// Place 3 ocean tiles owned by player
	positions := []shared.HexPosition{
		{Q: 0, R: 0, S: 0},
		{Q: 1, R: -1, S: 0},
		{Q: 2, R: -2, S: 0},
	}
	for _, pos := range positions {
		err := g.Board().UpdateTileOccupancy(ctx, pos, board.TileOccupant{
			Type: shared.ResourceOceanTile,
		}, playerID)
		if err != nil {
			t.Fatalf("Failed to place ocean: %v", err)
		}
	}

	score := gamecards.CalculateAwardScore(shared.AwardLandlord, p, g.Board(), cardRegistry)

	if score != 0 {
		t.Fatalf("expected Landlord score=0 with only oceans, got %d", score)
	}
}
