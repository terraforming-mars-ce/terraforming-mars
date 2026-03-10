package board_test

import (
	"testing"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
)

func TestGenerateMarsBoard_NoctisCityTaggedTile(t *testing.T) {
	tiles := board.GenerateMarsBoard(false)

	noctisCityPos := shared.HexPosition{Q: -2, R: 0, S: 2}

	var noctisTile *board.Tile
	for i := range tiles {
		if tiles[i].Coordinates.Equals(noctisCityPos) {
			noctisTile = &tiles[i]
			break
		}
	}

	if noctisTile == nil {
		t.Fatal("expected to find Noctis City tile at position {-2, 0, 2}")
	}

	// Check the tile has the noctis-city tag
	hasNoctisTag := false
	for _, tag := range noctisTile.Tags {
		if tag == board.BoardTagNoctisCity {
			hasNoctisTag = true
			break
		}
	}
	if !hasNoctisTag {
		t.Errorf("expected Noctis City tile to have tag %q, got tags: %v", board.BoardTagNoctisCity, noctisTile.Tags)
	}

	// Check the tile has a display name
	if noctisTile.DisplayName == nil {
		t.Fatal("expected Noctis City tile to have a display name")
	}
	if *noctisTile.DisplayName != "Noctis City" {
		t.Errorf("expected display name 'Noctis City', got %q", *noctisTile.DisplayName)
	}

	// Check it's a land tile, not ocean
	if noctisTile.Type != shared.ResourceLandTile {
		t.Errorf("expected Noctis City to be a land tile, got %s", noctisTile.Type)
	}
}

func TestGenerateMarsBoard_UntaggedTilesHaveNoTags(t *testing.T) {
	tiles := board.GenerateMarsBoard(false)

	// Count tiles with tags and without tags
	taggedCount := 0
	untaggedCount := 0

	for _, tile := range tiles {
		if len(tile.Tags) > 0 {
			taggedCount++
		} else {
			untaggedCount++
		}
	}

	// Noctis City, 4 volcanic tiles, and 3 off-Mars tiles (Phobos, Dawn City, Ganymede Colony) should be tagged
	if taggedCount != 8 {
		t.Errorf("expected exactly 8 tagged tiles (Noctis City, 4 volcanic, 3 off-Mars), got %d", taggedCount)
	}

	// All other tiles should have empty tags
	if untaggedCount == 0 {
		t.Error("expected some untagged tiles")
	}
}

func TestGenerateMarsBoard_TotalTileCount(t *testing.T) {
	tiles := board.GenerateMarsBoard(false)

	// Standard Mars board with radius 4 has 61 hex tiles + 3 off-Mars tiles = 64
	expectedCount := 64
	if len(tiles) != expectedCount {
		t.Errorf("expected %d tiles, got %d", expectedCount, len(tiles))
	}
}
