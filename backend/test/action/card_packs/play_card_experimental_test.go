package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Nuke (EXP004) ---
// "Replace any non-ocean tile with a Nuclear Zone tile."
// Uses tile-replacement:nuclear-zone, -2 VP, targets non-ocean occupied tiles only.

func TestNuke_ReplacesOccupiedTileWithNuclearZone(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// Place a city for p2 on a land tile
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	// Verify city is placed
	tile := testutil.GetTileAtHex(t, testGame, cityHex)
	testutil.AssertTrue(t, tile.OccupiedBy != nil, "City should be placed")
	testutil.AssertEqual(t, shared.ResourceCityTile, tile.OccupiedBy.Type, "Should be a city tile")

	// Play Nuke card
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: nukeCard.Cost}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke card should play successfully")

	// Should have pending tile selection for tile-replacement
	pending := testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "tile-replacement:nuclear-zone", pending.TileType, "Pending type should be tile-replacement:nuclear-zone")

	// The city hex should be in available hexes
	testutil.AssertTrue(t, testutil.ContainsHex(pending.AvailableHexes, cityHex),
		"City hex should be available for replacement")

	// Select the city hex for replacement
	selectTile := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTile.Execute(ctx, testGame.ID(), p1ID, cityHex)
	testutil.AssertNoError(t, err, "Tile replacement should succeed")

	// Verify the tile is now a nuclear zone
	tile = testutil.GetTileAtHex(t, testGame, cityHex)
	testutil.AssertTrue(t, tile.OccupiedBy != nil, "Tile should still be occupied")
	testutil.AssertEqual(t, shared.ResourceNuclearZoneTile, tile.OccupiedBy.Type, "Should be a nuclear zone tile")

	// Pending tile selection should be cleared
	pending = testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending == nil, "Pending tile selection should be cleared")
}

func TestNuke_DoesNotTargetOceanTiles(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// Place an ocean tile
	oceanHex := testutil.FindUnoccupiedOceanHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "ocean", oceanHex)

	// Place a city on a land tile
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	// Play Nuke card
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: nukeCard.Cost}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke card should play successfully")

	pending := testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending != nil, "Should have pending tile selection")

	// Ocean hex should NOT be in available hexes
	testutil.AssertTrue(t, !testutil.ContainsHex(pending.AvailableHexes, oceanHex),
		"Ocean hex should NOT be available for replacement")

	// City hex should be in available hexes
	testutil.AssertTrue(t, testutil.ContainsHex(pending.AvailableHexes, cityHex),
		"City hex should be available for replacement")
}

func TestNuke_HasNegativeVP(t *testing.T) {
	nukeCard := testutil.GetCardByName("Nuke")
	testutil.AssertTrue(t, len(nukeCard.VPConditions) > 0, "Nuke should have VP conditions")
	testutil.AssertEqual(t, -2, nukeCard.VPConditions[0].Amount, "Nuke should have -2 VP")
	testutil.AssertEqual(t, "fixed", nukeCard.VPConditions[0].Condition, "VP condition should be fixed")
}

func TestNuke_UnavailableWhenNoOccupiedNonOceanTiles(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)

	p1ID := playerIDs[0]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// No tiles placed — tile-replacement should find 0 available hexes
	availableHexes := testGame.CountAvailableHexesForTile("tile-replacement:nuclear-zone", p1ID, nil)
	testutil.AssertEqual(t, 0, availableHexes, "Should have no available hexes when no occupied non-ocean tiles")

	// Place only an ocean tile
	oceanHex := testutil.FindUnoccupiedOceanHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p1ID, "ocean", oceanHex)

	availableHexes = testGame.CountAvailableHexesForTile("tile-replacement:nuclear-zone", p1ID, nil)
	testutil.AssertEqual(t, 0, availableHexes, "Should have no available hexes when only ocean tiles exist")
}
