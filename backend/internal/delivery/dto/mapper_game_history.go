package dto

import (
	"terraforming-mars-backend/internal/game/datastore"
)

// ToGameHistoryEntryDtos converts a slice of history entries to DTOs.
func ToGameHistoryEntryDtos(entries []*datastore.GameStateHistoryEntry) []GameHistoryEntryDto {
	dtos := make([]GameHistoryEntryDto, len(entries))
	for i, entry := range entries {
		dtos[i] = toGameHistoryEntryDto(entry)
	}
	return dtos
}

func toGameHistoryEntryDto(entry *datastore.GameStateHistoryEntry) GameHistoryEntryDto {
	s := entry.State

	tileDtos := make([]TileDto, len(s.Tiles))
	for i, tile := range s.Tiles {
		tileDtos[i] = TileDto{
			Coordinates: HexPositionDto{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Type:        string(tile.Type),
			OwnerID:     tile.OwnerID,
			Tags:        tile.Tags,
			Bonuses:     convertTileBonuses(tile.Bonuses),
			Location:    string(tile.Location),
			DisplayName: tile.DisplayName,
			ReservedBy:  tile.ReservedBy,
		}
		if tile.OccupiedBy != nil {
			tileDtos[i].OccupiedBy = &TileOccupantDto{
				Type: string(tile.OccupiedBy.Type),
				Tags: tile.OccupiedBy.Tags,
			}
		}
	}

	players := make(map[string]GameHistoryPlayerDto, len(s.Players))
	for id, p := range s.Players {
		// totalVP comes from the snapshot enricher in main.go, which calls
		// gameAction.ComputePlayerVPBreakdowns — the same helper FinalScoringAction
		// uses. If the enricher wasn't wired (legacy entries, tests), totalVP is 0.
		totalVP := 0
		if entry.VPBreakdowns != nil {
			if breakdown, ok := entry.VPBreakdowns[id]; ok {
				totalVP = breakdown.TotalVP
			}
		}

		players[id] = GameHistoryPlayerDto{
			ID:              p.ID,
			Name:            p.Name,
			Color:           p.Color,
			TerraformRating: p.TerraformRating,
			Credits:         p.Resources.Credits,
			Steel:           p.Resources.Steel,
			Titanium:        p.Resources.Titanium,
			Plants:          p.Resources.Plants,
			Energy:          p.Resources.Energy,
			Heat:            p.Resources.Heat,
			PlayedCardCount: len(p.PlayedCardIDs),
			Production:      toProductionDto(p.Production),
			PlayedCardIDs:   p.PlayedCardIDs,
			HandCardIDs:     p.HandCardIDs,
			CorporationID:   p.CorporationID,
			ResourceStorage: p.ResourceStorage,
			TotalVP:         totalVP,
		}
	}

	milestones := make([]ClaimedMilestoneDto, len(s.ClaimedMilestones))
	for i, m := range s.ClaimedMilestones {
		milestones[i] = ClaimedMilestoneDto{
			Type:     string(m.Type),
			PlayerID: m.PlayerID,
		}
	}

	awards := make([]FundedAwardDto, len(s.FundedAwards))
	for i, a := range s.FundedAwards {
		awards[i] = FundedAwardDto{
			Type:     string(a.Type),
			PlayerID: a.FundedByPlayer,
		}
	}

	return GameHistoryEntryDto{
		Sequence:     entry.Sequence,
		Timestamp:    entry.Timestamp,
		Generation:   s.Generation,
		Phase:        GamePhase(s.CurrentPhase),
		Temperature:  s.Temperature,
		Oxygen:       s.Oxygen,
		Oceans:       s.Oceans,
		Venus:        s.Venus,
		ActionNumber: s.GlobalActionCounter,
		Board:        BoardDto{Tiles: tileDtos},
		Players:      players,
		Milestones:   milestones,
		Awards:       awards,
		Settings: GameSettingsDto{
			MaxPlayers:       s.Settings.MaxPlayers,
			VenusNextEnabled: s.Settings.VenusNextEnabled,
			DevelopmentMode:  s.Settings.DevelopmentMode,
			DemoGame:         s.Settings.DemoGame,
			CardPacks:        s.Settings.CardPacks,
		},
	}
}
