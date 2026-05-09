package dto_test

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestToGameHistoryEntryDtos_TotalVP_FromBreakdown(t *testing.T) {
	playerID := "player-1"

	entry := &datastore.GameStateHistoryEntry{
		GameID:    "game-1",
		Sequence:  1,
		Timestamp: time.Now(),
		State: &datastore.GameState{
			Players: map[string]*datastore.PlayerState{
				playerID: {
					ID:              playerID,
					Name:            "Test Player",
					Color:           "red",
					TerraformRating: 25,
					Resources:       shared.Resources{},
					ResourceStorage: map[string]int{},
				},
			},
			ClaimedMilestones: []shared.ClaimedMilestone{},
			FundedAwards:      []shared.FundedAward{},
			Settings:          shared.GameSettings{},
		},
		VPBreakdowns: map[string]shared.VPBreakdown{
			playerID: {
				TerraformRating: 25,
				CardVP:          7,
				MilestoneVP:     5,
				AwardVP:         5,
				GreeneryVP:      2,
				CityVP:          2,
				TotalVP:         46,
			},
		},
	}

	dtos := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{entry})
	testutil.AssertEqual(t, 1, len(dtos), "Should have one entry")
	testutil.AssertEqual(t, 46, dtos[0].Players[playerID].TotalVP, "TotalVP should come from VPBreakdowns")
}

func TestToGameHistoryEntryDtos_TotalVP_NilBreakdownDegradesGracefully(t *testing.T) {
	playerID := "player-1"

	entry := &datastore.GameStateHistoryEntry{
		GameID:    "game-1",
		Sequence:  1,
		Timestamp: time.Now(),
		State: &datastore.GameState{
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
		// VPBreakdowns deliberately nil — represents legacy entries or tests
		// where the snapshot enricher wasn't wired. Mapper should not panic
		// and should report 0.
	}

	dtos := dto.ToGameHistoryEntryDtos([]*datastore.GameStateHistoryEntry{entry})
	testutil.AssertEqual(t, 0, dtos[0].Players[playerID].TotalVP, "TotalVP should be 0 when VPBreakdowns is nil")
}
