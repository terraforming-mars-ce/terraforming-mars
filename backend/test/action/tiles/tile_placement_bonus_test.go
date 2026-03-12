package tiles_test

import (
	"context"
	"fmt"
	"testing"

	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestSelectTileAction_BonusesRemovedAfterClaim(t *testing.T) {
	ctx := context.Background()
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	// Start the game to get to action phase
	testutil.StartTestGame(t, testGame)

	playerID := testGame.TurnOrder()[0]
	p, _ := testGame.GetPlayer(playerID)

	// Give player enough credits
	testutil.SetPlayerCredits(ctx, p, 100)

	// Find a tile with bonuses on the board
	tiles := testGame.Board().Tiles()
	var bonusTileCoords *shared.HexPosition
	var expectedBonus board.TileBonus
	for _, tile := range tiles {
		if len(tile.Bonuses) > 0 && tile.OccupiedBy == nil && tile.Type == shared.ResourceLandTile {
			bonusTileCoords = &tile.Coordinates
			expectedBonus = tile.Bonuses[0]
			break
		}
	}

	if bonusTileCoords == nil {
		t.Fatal("No tile with bonuses found on board")
	}

	// Record initial resources
	initialResources := getResourceAmount(p, expectedBonus.Type)

	// Set up pending tile selection for the player (simulating city placement)
	hexStr := formatHexCoords(*bonusTileCoords)
	pendingSelection := &shared.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, pendingSelection), "set pending tile selection")

	// Execute tile placement
	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err := selectTileAction.Execute(ctx, testGame.ID(), playerID, hexStr)
	testutil.AssertNoError(t, err, "Failed to place tile")

	// Verify bonus was awarded to player
	finalResources := getResourceAmount(p, expectedBonus.Type)
	expectedFinal := initialResources + expectedBonus.Amount
	testutil.AssertEqual(t, expectedFinal, finalResources,
		"Player should have received bonus resources")

	// Verify bonus was removed from the tile
	placedTile, err := testGame.Board().GetTile(*bonusTileCoords)
	testutil.AssertNoError(t, err, "Failed to get placed tile")
	testutil.AssertEqual(t, 0, len(placedTile.Bonuses),
		"Tile bonuses should be cleared after placement")
}

func TestSelectTileAction_CardDrawBonusRemovedAfterClaim(t *testing.T) {
	ctx := context.Background()
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	// Start the game
	testutil.StartTestGame(t, testGame)

	playerID := testGame.TurnOrder()[0]
	p, _ := testGame.GetPlayer(playerID)

	// Find a tile with card draw bonus
	tiles := testGame.Board().Tiles()
	var cardDrawTileCoords *shared.HexPosition
	for _, tile := range tiles {
		for _, bonus := range tile.Bonuses {
			if bonus.Type == shared.ResourceCardDraw && tile.OccupiedBy == nil && tile.Type == shared.ResourceLandTile {
				cardDrawTileCoords = &tile.Coordinates
				break
			}
		}
		if cardDrawTileCoords != nil {
			break
		}
	}

	if cardDrawTileCoords == nil {
		t.Fatal("No tile with card draw bonus found on board")
	}

	// Record initial hand size
	initialHandSize := len(p.Hand().Cards())

	// Set up pending tile selection
	hexStr := formatHexCoords(*cardDrawTileCoords)
	pendingSelection := &shared.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, pendingSelection), "set pending tile selection")

	// Execute tile placement
	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err := selectTileAction.Execute(ctx, testGame.ID(), playerID, hexStr)
	testutil.AssertNoError(t, err, "Failed to place tile")

	// Verify cards were drawn
	finalHandSize := len(p.Hand().Cards())
	testutil.AssertTrue(t, finalHandSize > initialHandSize,
		"Player should have received card draw bonus")

	// Verify bonus was removed from the tile
	placedTile, err := testGame.Board().GetTile(*cardDrawTileCoords)
	testutil.AssertNoError(t, err, "Failed to get placed tile")
	testutil.AssertEqual(t, 0, len(placedTile.Bonuses),
		"Card draw bonus should be cleared after placement")
}

func TestSelectTileAction_OceanAdjacencyBonus(t *testing.T) {
	ctx := context.Background()
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	testutil.StartTestGame(t, testGame)

	playerID := testGame.TurnOrder()[0]
	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Place an ocean tile at (1,1,-2) directly on the board
	oceanCoords := shared.HexPosition{Q: 1, R: 1, S: -2}
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanCoords,
		board.TileOccupant{Type: shared.ResourceOceanTile}, playerID)
	testutil.AssertNoError(t, err, "Should place ocean tile")

	// Record credits before placing the adjacent city
	initialCredits := p.Resources().Get().Credits

	// Place a city adjacent to the ocean at (2,0,-2) — neighbor of (1,1,-2)
	cityCoords := shared.HexPosition{Q: 2, R: 0, S: -2}
	hexStr := formatHexCoords(cityCoords)
	pendingSelection := &shared.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, pendingSelection), "set pending tile selection")

	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTileAction.Execute(ctx, testGame.ID(), playerID, hexStr)
	testutil.AssertNoError(t, err, "Failed to place city tile")

	// Verify player received +2 M€ for one adjacent ocean
	finalCredits := p.Resources().Get().Credits
	testutil.AssertEqual(t, initialCredits+2, finalCredits,
		"Player should receive +2 M€ for one adjacent ocean tile")
}

func TestSelectTileAction_OceanAdjacencyBonus_MultipleOceans(t *testing.T) {
	ctx := context.Background()
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	testutil.StartTestGame(t, testGame)

	playerID := testGame.TurnOrder()[0]
	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Place two ocean tiles adjacent to the same land tile (1,2,-3)
	// Ocean at (1,1,-2) — neighbor via East direction
	ocean1 := shared.HexPosition{Q: 1, R: 1, S: -2}
	err := testGame.Board().UpdateTileOccupancy(ctx, ocean1,
		board.TileOccupant{Type: shared.ResourceOceanTile}, playerID)
	testutil.AssertNoError(t, err, "Should place first ocean tile")

	// Ocean at (0,3,-3) — neighbor via West direction
	ocean2 := shared.HexPosition{Q: 0, R: 3, S: -3}
	err = testGame.Board().UpdateTileOccupancy(ctx, ocean2,
		board.TileOccupant{Type: shared.ResourceOceanTile}, playerID)
	testutil.AssertNoError(t, err, "Should place second ocean tile")

	initialCredits := p.Resources().Get().Credits

	// Place a greenery at (1,2,-3) — adjacent to both oceans
	greeneryCoords := shared.HexPosition{Q: 1, R: 2, S: -3}
	hexStr := formatHexCoords(greeneryCoords)
	pendingSelection := &shared.PendingTileSelection{
		TileType:       "greenery",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, pendingSelection), "set pending tile selection")

	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTileAction.Execute(ctx, testGame.ID(), playerID, hexStr)
	testutil.AssertNoError(t, err, "Failed to place greenery tile")

	// Verify player received +4 M€ for two adjacent oceans
	finalCredits := p.Resources().Get().Credits
	testutil.AssertEqual(t, initialCredits+4, finalCredits,
		"Player should receive +4 M€ for two adjacent ocean tiles")
}

func TestSelectTileAction_OceanAdjacencyBonus_OceanTileGetsBonus(t *testing.T) {
	ctx := context.Background()
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	testutil.StartTestGame(t, testGame)

	playerID := testGame.TurnOrder()[0]
	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Place an ocean tile at (2,-1,-1)
	ocean1 := shared.HexPosition{Q: 2, R: -1, S: -1}
	err := testGame.Board().UpdateTileOccupancy(ctx, ocean1,
		board.TileOccupant{Type: shared.ResourceOceanTile}, playerID)
	testutil.AssertNoError(t, err, "Should place first ocean tile")

	initialCredits := p.Resources().Get().Credits

	// Place another ocean adjacent to the first at (3,-2,-1) — also an ocean space, neighbor via East
	ocean2Coords := shared.HexPosition{Q: 3, R: -2, S: -1}
	hexStr := formatHexCoords(ocean2Coords)
	pendingSelection := &shared.PendingTileSelection{
		TileType:       "ocean",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, pendingSelection), "set pending tile selection")

	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTileAction.Execute(ctx, testGame.ID(), playerID, hexStr)
	testutil.AssertNoError(t, err, "Failed to place ocean tile")

	// Ocean tiles SHOULD receive adjacency bonus (2 MC per adjacent ocean)
	finalCredits := p.Resources().Get().Credits
	testutil.AssertEqual(t, initialCredits+2, finalCredits,
		"Ocean tile placement should receive +2 M€ for one adjacent ocean tile")
}

func getResourceAmount(p *player.Player, resourceType shared.ResourceType) int {
	resources := p.Resources().Get()
	switch resourceType {
	case shared.ResourceSteel:
		return resources.Steel
	case shared.ResourceTitanium:
		return resources.Titanium
	case shared.ResourcePlant:
		return resources.Plants
	case shared.ResourceCredit:
		return resources.Credits
	default:
		return 0
	}
}

func formatHexCoords(pos shared.HexPosition) string {
	return fmt.Sprintf("%d,%d,%d", pos.Q, pos.R, pos.S)
}
