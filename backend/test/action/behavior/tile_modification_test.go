package behavior_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestTileModification_DestructionCreatesPendingSelection(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)
	ctx := context.Background()

	// Place a tile on the board so there is something to destroy
	tiles := testGame.Board().Tiles()
	for _, tile := range tiles {
		if tile.Location == board.TileLocationMars && tile.OccupiedBy == nil {
			err := testGame.Board().UpdateTileOccupancy(ctx, tile.Coordinates,
				board.TileOccupant{Type: shared.ResourceGreeneryTile}, playerID)
			testutil.AssertNoError(t, err, "placing greenery tile")
			break
		}
	}

	output := shared.NewTileModificationCondition(shared.ResourceTileDestruction, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	// The queue is immediately consumed via ProcessNextTile, creating a PendingTileSelection
	sel := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, sel != nil, "pending tile selection should exist")
	testutil.AssertEqual(t, "tile-destruction", sel.TileType, "tile type should be tile-destruction")
	testutil.AssertTrue(t, len(sel.AvailableHexes) > 0, "should have available hexes for destruction")
}

func TestTileModification_ReplacementWithTileType(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)
	ctx := context.Background()

	// Place a tile on the board so there is something to replace
	tiles := testGame.Board().Tiles()
	for _, tile := range tiles {
		if tile.Location == board.TileLocationMars && tile.OccupiedBy == nil {
			err := testGame.Board().UpdateTileOccupancy(ctx, tile.Coordinates,
				board.TileOccupant{Type: shared.ResourceGreeneryTile}, playerID)
			testutil.AssertNoError(t, err, "placing greenery tile")
			break
		}
	}

	output := &shared.TileModificationCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTileReplacement, Amount: 1, Target: "none"},
		TileType:      "world-tree",
	}
	applyOutputs(t, p, testGame, cardRegistry, output)

	// The queue is immediately consumed via ProcessNextTile, creating a PendingTileSelection
	sel := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, sel != nil, "pending tile selection should exist")
	testutil.AssertEqual(t, "tile-replacement:world-tree", sel.TileType, "tile type should be tile-replacement:world-tree")
	testutil.AssertTrue(t, len(sel.AvailableHexes) > 0, "should have available hexes for replacement")
}
