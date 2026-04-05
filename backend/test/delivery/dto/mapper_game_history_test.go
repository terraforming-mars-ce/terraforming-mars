package dto_test

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestToGameHistoryEntryDtos_TotalVP(t *testing.T) {
	playerID := "player-1"
	ownerID := playerID

	tiles := []board.Tile{
		{
			Coordinates: shared.HexPosition{Q: 0, R: 0, S: 0},
			Type:        shared.ResourceLandTile,
			OccupiedBy:  &board.TileOccupant{Type: shared.ResourceCityTile},
			OwnerID:     &ownerID,
		},
		{
			Coordinates: shared.HexPosition{Q: 1, R: -1, S: 0},
			Type:        shared.ResourceLandTile,
			OccupiedBy:  &board.TileOccupant{Type: shared.ResourceGreeneryTile},
			OwnerID:     &ownerID,
		},
		{
			Coordinates: shared.HexPosition{Q: 0, R: 1, S: -1},
			Type:        shared.ResourceLandTile,
			OccupiedBy:  &board.TileOccupant{Type: shared.ResourceGreeneryTile},
			OwnerID:     &ownerID,
		},
	}

	entry := &datastore.GameStateHistoryEntry{
		GameID:    "game-1",
		Sequence:  1,
		Timestamp: time.Now(),
		State: &datastore.GameState{
			Tiles: tiles,
			Players: map[string]*datastore.PlayerState{
				playerID: {
					ID:              playerID,
					Name:            "Test Player",
					Color:           "red",
					TerraformRating: 25,
					VPGranters: []shared.VPGranter{
						{CardID: "card-1", CardName: "Card One", ComputedValue: 3},
						{CardID: "card-2", CardName: "Card Two", ComputedValue: 4},
					},
					Resources:       shared.Resources{},
					ResourceStorage: map[string]int{},
				},
			},
			ClaimedMilestones: []shared.ClaimedMilestone{
				{Type: "terraformer", PlayerID: playerID},
			},
			FundedAwards: []shared.FundedAward{},
			Settings:     shared.GameSettings{},
		},
	}

	dtos := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{entry})
	testutil.AssertEqual(t, 1, len(dtos), "Should have one entry")

	player := dtos[0].Players[playerID]
	// TR=25, cardVP=3+4=7, greeneryVP=2, cityVP=2 (city at 0,0,0 adjacent to 2 greeneries), milestoneVP=5
	expectedVP := 25 + 7 + 2 + 2 + 5
	testutil.AssertEqual(t, expectedVP, player.TotalVP, "TotalVP should include TR, card VP, greenery VP, city VP, and milestone VP")
}

func TestToGameHistoryEntryDtos_TotalVP_NoVPGranters(t *testing.T) {
	playerID := "player-1"

	entry := &datastore.GameStateHistoryEntry{
		GameID:    "game-1",
		Sequence:  1,
		Timestamp: time.Now(),
		State: &datastore.GameState{
			Tiles: []board.Tile{},
			Players: map[string]*datastore.PlayerState{
				playerID: {
					ID:              playerID,
					Name:            "Test Player",
					Color:           "blue",
					TerraformRating: 20,
					Resources:       shared.Resources{},
					ResourceStorage: map[string]int{},
				},
			},
			ClaimedMilestones: []shared.ClaimedMilestone{},
			FundedAwards:      []shared.FundedAward{},
			Settings:          shared.GameSettings{},
		},
	}

	dtos := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{entry})
	player := dtos[0].Players[playerID]
	testutil.AssertEqual(t, 20, player.TotalVP, "TotalVP should equal TR when no other VP sources exist")
}

func TestToGameHistoryEntryDtos_TotalVP_WorldTreeCountsAsGreenery(t *testing.T) {
	playerID := "player-1"
	ownerID := playerID

	tiles := []board.Tile{
		{
			Coordinates: shared.HexPosition{Q: 0, R: 0, S: 0},
			Type:        shared.ResourceLandTile,
			OccupiedBy:  &board.TileOccupant{Type: shared.ResourceWorldTreeTile},
			OwnerID:     &ownerID,
		},
	}

	entry := &datastore.GameStateHistoryEntry{
		GameID:    "game-1",
		Sequence:  1,
		Timestamp: time.Now(),
		State: &datastore.GameState{
			Tiles: tiles,
			Players: map[string]*datastore.PlayerState{
				playerID: {
					ID:              playerID,
					Name:            "Test Player",
					Color:           "green",
					TerraformRating: 20,
					Resources:       shared.Resources{},
					ResourceStorage: map[string]int{},
				},
			},
			ClaimedMilestones: []shared.ClaimedMilestone{},
			FundedAwards:      []shared.FundedAward{},
			Settings:          shared.GameSettings{},
		},
	}

	dtos := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{entry})
	player := dtos[0].Players[playerID]
	// TR=20, greeneryVP=1 (world-tree counts as greenery)
	testutil.AssertEqual(t, 21, player.TotalVP, "World-tree tile should count as greenery for VP")
}
