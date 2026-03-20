package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestWorldTreeVP_AdjacentGreeneries verifies that the World Tree card's VP
// is computed via the generic VPCondition system using adjacentToTileType,
// not via any hardcoded function.
func TestWorldTreeVP_AdjacentGreeneries(t *testing.T) {
	g, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	cardID := testutil.CardID("The World Tree")
	p.PlayedCards().AddCard(cardID, "The World Tree", "automated", []string{"plant", "building"})

	// Place a world-tree tile at (0,0,0) owned by the player
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place world-tree tile: %v", err)
	}

	// Place 2 greenery tiles adjacent to world-tree
	greeneryPositions := []shared.HexPosition{
		{Q: 1, R: -1, S: 0}, // neighbor of (0,0,0)
		{Q: 0, R: -1, S: 1}, // neighbor of (0,0,0)
	}
	for _, pos := range greeneryPositions {
		err := g.Board().UpdateTileOccupancy(ctx, pos, board.TileOccupant{
			Type: shared.ResourceGreeneryTile,
		}, "other-player")
		if err != nil {
			t.Fatalf("failed to place greenery: %v", err)
		}
	}

	// Place a greenery NOT adjacent to world-tree (should not count)
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 3, R: -3, S: 0}, board.TileOccupant{
		Type: shared.ResourceGreeneryTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place non-adjacent greenery: %v", err)
	}

	breakdown := gamecards.CalculatePlayerVP(
		p,
		g.Board(),
		nil,
		nil,
		g.GetAllPlayers(),
		cardRegistry,
		nil,
	)

	// World Tree should give 2 VP (2 adjacent greeneries)
	if breakdown.CardVP != 2 {
		t.Fatalf("expected CardVP=2 from World Tree with 2 adjacent greeneries, got %d", breakdown.CardVP)
	}

	// Verify the VP comes from the card detail
	found := false
	for _, detail := range breakdown.CardVPDetails {
		if detail.CardID == cardID {
			found = true
			if detail.TotalVP != 2 {
				t.Errorf("expected World Tree card detail TotalVP=2, got %d", detail.TotalVP)
			}
		}
	}
	if !found {
		t.Error("World Tree card not found in CardVPDetails")
	}
}

// TestWorldTreeVP_AdjacentWorldTreeCountsAsForest verifies that another
// world-tree tile adjacent to the player's world-tree counts as a forest.
func TestWorldTreeVP_AdjacentWorldTreeCountsAsForest(t *testing.T) {
	g, _, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	cardID := testutil.CardID("The World Tree")
	p.PlayedCards().AddCard(cardID, "The World Tree", "automated", []string{"plant", "building"})

	// Place world-tree at (0,0,0) owned by the player
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place world-tree tile: %v", err)
	}

	// Place another world-tree adjacent (owned by other player)
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 1, R: -1, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, otherPlayerID)
	if err != nil {
		t.Fatalf("failed to place adjacent world-tree: %v", err)
	}

	// Place a greenery adjacent
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: -1, S: 1}, board.TileOccupant{
		Type: shared.ResourceGreeneryTile,
	}, otherPlayerID)
	if err != nil {
		t.Fatalf("failed to place greenery: %v", err)
	}

	breakdown := gamecards.CalculatePlayerVP(
		p,
		g.Board(),
		nil,
		nil,
		g.GetAllPlayers(),
		cardRegistry,
		nil,
	)

	// Should be 2 VP: 1 for adjacent world-tree (counts as forest) + 1 for adjacent greenery
	if breakdown.CardVP != 2 {
		t.Fatalf("expected CardVP=2 (adjacent world-tree + greenery), got %d", breakdown.CardVP)
	}
}

// TestWorldTreeVP_NoAdjacentForests verifies 0 VP when no forests are adjacent.
func TestWorldTreeVP_NoAdjacentForests(t *testing.T) {
	g, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	cardID := testutil.CardID("The World Tree")
	p.PlayedCards().AddCard(cardID, "The World Tree", "automated", []string{"plant", "building"})

	// Place world-tree with no adjacent forests
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place world-tree tile: %v", err)
	}

	// Place a city adjacent (not a forest)
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 1, R: -1, S: 0}, board.TileOccupant{
		Type: shared.ResourceCityTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place city: %v", err)
	}

	breakdown := gamecards.CalculatePlayerVP(
		p,
		g.Board(),
		nil,
		nil,
		g.GetAllPlayers(),
		cardRegistry,
		nil,
	)

	if breakdown.CardVP != 0 {
		t.Fatalf("expected CardVP=0 with no adjacent forests, got %d", breakdown.CardVP)
	}
}

// TestWorldTreeVP_DeduplicatesSharedNeighbors verifies that a forest tile
// adjacent to multiple world-tree tiles is only counted once.
func TestWorldTreeVP_DeduplicatesSharedNeighbors(t *testing.T) {
	g, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)
	ctx := context.Background()

	// The player needs the card played to get VP from it
	cardID := testutil.CardID("The World Tree")
	p.PlayedCards().AddCard(cardID, "The World Tree", "automated", []string{"plant", "building"})

	// Place two world-tree tiles adjacent to each other, both owned by player
	err := g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: 0, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place world-tree 1: %v", err)
	}
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 1, R: -1, S: 0}, board.TileOccupant{
		Type: shared.ResourceWorldTreeTile,
	}, playerID)
	if err != nil {
		t.Fatalf("failed to place world-tree 2: %v", err)
	}

	// Place a greenery adjacent to BOTH world-trees
	// (0,-1,1) is neighbor of both (0,0,0) and (1,-1,0)
	err = g.Board().UpdateTileOccupancy(ctx, shared.HexPosition{Q: 0, R: -1, S: 1}, board.TileOccupant{
		Type: shared.ResourceGreeneryTile,
	}, "other")
	if err != nil {
		t.Fatalf("failed to place shared greenery: %v", err)
	}

	breakdown := gamecards.CalculatePlayerVP(
		p,
		g.Board(),
		nil,
		nil,
		g.GetAllPlayers(),
		cardRegistry,
		nil,
	)

	// The greenery should be counted once (deduplication), plus
	// the two world-trees are adjacent to each other and both count as forest.
	// World-tree at (0,0,0) neighbors: world-tree at (1,-1,0) [forest] + greenery at (0,-1,1) [forest] = 2
	// World-tree at (1,-1,0) neighbors: world-tree at (0,0,0) [forest] + greenery at (0,-1,1) [forest] = 2
	// But deduplication: unique adjacent forest tiles across all owned world-trees
	// Unique tiles: (1,-1,0) world-tree, (0,-1,1) greenery, (0,0,0) world-tree = 3
	// Wait - world-tree tiles themselves are adjacent to each other. Let's count:
	// From (0,0,0): adjacent forests = (1,-1,0) [world-tree] + (0,-1,1) [greenery]
	// From (1,-1,0): adjacent forests = (0,0,0) [world-tree] + (0,-1,1) [greenery]
	// Unique set: {(1,-1,0), (0,-1,1), (0,0,0)} = 3 unique tiles
	if breakdown.CardVP != 3 {
		t.Fatalf("expected CardVP=3 (deduplicated adjacent forests across 2 world-trees), got %d", breakdown.CardVP)
	}
}
