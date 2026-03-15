package testutil

import (
	"context"
	"fmt"
	"testing"

	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
)

// FormatHex formats a HexPosition as "q,r,s".
func FormatHex(pos shared.HexPosition) string {
	return fmt.Sprintf("%d,%d,%d", pos.Q, pos.R, pos.S)
}

// FindUnoccupiedLandHex returns the coordinate string of the first unoccupied, unreserved land hex.
func FindUnoccupiedLandHex(t *testing.T, g *game.Game) string {
	t.Helper()
	for _, tile := range g.Board().Tiles() {
		if tile.OccupiedBy == nil && tile.Type == shared.ResourceLandTile && tile.ReservedBy == nil {
			return FormatHex(tile.Coordinates)
		}
	}
	t.Fatal("No unoccupied land hex found")
	return ""
}

// FindUnoccupiedOceanHex returns the coordinate string of the first unoccupied, unreserved ocean space.
func FindUnoccupiedOceanHex(t *testing.T, g *game.Game) string {
	t.Helper()
	for _, tile := range g.Board().Tiles() {
		if tile.OccupiedBy == nil && tile.Type == shared.ResourceOceanSpace && tile.ReservedBy == nil {
			return FormatHex(tile.Coordinates)
		}
	}
	t.Fatal("No unoccupied ocean hex found")
	return ""
}

// GetTileAtHex returns the tile at the given "q,r,s" coordinate string.
func GetTileAtHex(t *testing.T, g *game.Game, hexStr string) board.Tile {
	t.Helper()
	for _, tile := range g.Board().Tiles() {
		if FormatHex(tile.Coordinates) == hexStr {
			return tile
		}
	}
	t.Fatalf("Tile not found at hex %s", hexStr)
	return board.Tile{}
}

// ContainsHex checks if a hex string is present in a slice.
func ContainsHex(hexes []string, hex string) bool {
	for _, h := range hexes {
		if h == hex {
			return true
		}
	}
	return false
}

// PlaceTileForPlayer sets up a pending tile selection and executes it to place a tile.
func PlaceTileForPlayer(ctx context.Context, t *testing.T, g *game.Game, repo game.GameRepository, playerID string, tileType string, hexStr string) {
	t.Helper()
	logger := TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	cr := CreateTestCardRegistry()

	err := g.SetCurrentTurn(ctx, playerID, 2)
	if err != nil {
		t.Fatalf("Failed to set turn for tile placement: %v", err)
	}

	err = g.SetPendingTileSelection(ctx, playerID, &shared.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: []string{hexStr},
		Source:         "test",
	})
	if err != nil {
		t.Fatalf("Failed to set pending tile selection: %v", err)
	}

	selectTile := tileAction.NewSelectTileAction(repo, cr, stateRepo, logger)
	_, err = selectTile.Execute(ctx, g.ID(), playerID, hexStr)
	if err != nil {
		t.Fatalf("Failed to place %s tile on %s: %v", tileType, hexStr, err)
	}
}
