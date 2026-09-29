package tiles_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestTileVisualPlacementAndViewerConsistency(t *testing.T) {
	ctx := context.Background()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, testutil.NewMockBroadcaster())
	testutil.StartTestGame(t, g)
	playerID := g.TurnOrder()[0]
	data, err := os.ReadFile("../../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatal(err)
	}
	var definitions []cards.Card
	if err = json.Unmarshal(data, &definitions); err != nil {
		t.Fatal(err)
	}
	registry := cards.NewInMemoryCardRegistry(definitions)
	for _, id := range []string{"020", "008", "017", "029", "120", "032", "108"} {
		card, err := registry.GetByID(id)
		if err != nil || card.Style == nil || card.Style.Tile == nil {
			t.Fatalf("Missing card style %s: %v", id, err)
		}
		if dto.ToCardDto(*card).Style.Tile == nil {
			t.Fatal("Card DTO lost style")
		}
	}
	var coordinate shared.HexPosition
	found := false
	for _, tile := range g.Board().Tiles() {
		if tile.Type == shared.ResourceLandTile && tile.OccupiedBy == nil && len(tile.Bonuses) == 0 {
			coordinate = tile.Coordinates
			found = true
			break
		}
	}
	if !found {
		t.Fatal("No empty land tile")
	}
	if err = g.SetPendingTileSelection(ctx, playerID, &shared.PendingTileSelection{TileType: "city", AvailableHexes: []string{coordinate.String()}, Source: "card", SourceCardID: "032"}); err != nil {
		t.Fatal(err)
	}
	action := tileAction.NewSelectTileAction(repo, registry, game.NewInMemoryGameStateRepository(), testutil.TestLogger())
	if _, err = action.Execute(ctx, g.ID(), playerID, coordinate.String()); err != nil {
		t.Fatal(err)
	}
	placed, err := g.Board().GetTile(coordinate)
	if err != nil {
		t.Fatal(err)
	}
	visual := placed.OccupiedBy.Visual
	if visual == nil || visual.Seed == 0 || visual.City == nil || visual.City.Cover != "flat-glass" {
		t.Fatalf("Invalid visual: %+v", visual)
	}
	expected := visual.Clone()
	visual.City.Details[0] = "mutated"
	again, _ := g.Board().GetTile(coordinate)
	if !reflect.DeepEqual(again.OccupiedBy.Visual, expected) {
		t.Fatal("Board read mutated saved appearance")
	}
	card, _ := registry.GetByID("032")
	card.Style.Tile.Cover = "dome"
	again, _ = g.Board().GetTile(coordinate)
	if again.OccupiedBy.Visual.City.Cover != "flat-glass" {
		t.Fatal("Card request aliases placed appearance")
	}
	playerView := dto.ToGameDto(g, registry, playerID)
	spectatorView := dto.ToSpectatorGameDto(g, registry, nil, nil)
	history := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{{State: &datastore.GameState{Tiles: g.Board().Tiles()}}})
	var views []*dto.TileVisualDto
	for _, tiles := range [][]dto.TileDto{playerView.Board.Tiles, spectatorView.Board.Tiles, history[0].Board.Tiles} {
		for _, tile := range tiles {
			if tile.Coordinates.Q == coordinate.Q && tile.Coordinates.R == coordinate.R {
				views = append(views, tile.OccupiedBy.Visual)
			}
		}
	}
	if len(views) != 3 || !reflect.DeepEqual(views[0], views[1]) || !reflect.DeepEqual(views[0], views[2]) {
		t.Fatal("Viewers disagree on visual specification")
	}
	if err = g.Board().ClearTileOccupant(ctx, coordinate); err != nil {
		t.Fatal(err)
	}
	cleared, _ := g.Board().GetTile(coordinate)
	if cleared.OccupiedBy != nil {
		t.Fatal("Cleared tile kept appearance")
	}
	if err = g.Board().UpdateTileOccupancy(ctx, coordinate, board.TileOccupant{Type: shared.ResourceCityTile}, playerID); err != nil {
		t.Fatal(err)
	}
	standard, _ := g.Board().GetTile(coordinate)
	if standard.OccupiedBy.Visual == nil || standard.OccupiedBy.Visual.City != nil {
		t.Fatal("Standard city retained card request")
	}
	seed := standard.OccupiedBy.Visual.Seed
	if err = g.Board().UpdateTileOccupancy(ctx, coordinate, board.TileOccupant{Type: shared.ResourceCityTile}, playerID); err != nil {
		t.Fatal(err)
	}
	repeated, _ := g.Board().GetTile(coordinate)
	if repeated.OccupiedBy.Visual.Seed != seed {
		t.Fatal("Seed is not deterministic")
	}
}
