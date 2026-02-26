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
	pendingSelection := &player.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testGame.SetPendingTileSelection(ctx, playerID, pendingSelection)

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
	pendingSelection := &player.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{hexStr},
		Source:         "test",
	}
	testGame.SetPendingTileSelection(ctx, playerID, pendingSelection)

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
