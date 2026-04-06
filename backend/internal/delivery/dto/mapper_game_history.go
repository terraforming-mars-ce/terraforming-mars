package dto

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
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
		cardVP := computeHistoryCardVP(p)
		greeneryVP := computeHistoryGreeneryVP(s.Tiles, id)
		cityVP := computeHistoryCityVP(s.Tiles, id)
		milestoneVP := computeHistoryMilestoneVP(s.ClaimedMilestones, id)
		totalVP := p.TerraformRating + cardVP + greeneryVP + cityVP + milestoneVP

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

func computeHistoryCardVP(p *datastore.PlayerState) int {
	vp := 0
	for _, g := range p.VPGranters {
		vp += g.ComputedValue
	}
	return vp
}

func computeHistoryGreeneryVP(tiles []board.Tile, playerID string) int {
	vp := 0
	for _, t := range tiles {
		if t.OccupiedBy != nil && shared.IsForestTile(t.OccupiedBy.Type) && t.OwnerID != nil && *t.OwnerID == playerID {
			vp++
		}
	}
	return vp
}

func computeHistoryCityVP(tiles []board.Tile, playerID string) int {
	vp := 0
	for _, t := range tiles {
		if t.OccupiedBy == nil || t.OccupiedBy.Type != shared.ResourceCityTile || t.OwnerID == nil || *t.OwnerID != playerID {
			continue
		}
		for _, neighbor := range t.Coordinates.GetNeighbors() {
			for _, n := range tiles {
				if n.Coordinates == neighbor && n.OccupiedBy != nil && shared.IsForestTile(n.OccupiedBy.Type) {
					vp++
					break
				}
			}
		}
	}
	return vp
}

func computeHistoryMilestoneVP(milestones []shared.ClaimedMilestone, playerID string) int {
	vp := 0
	for _, m := range milestones {
		if m.PlayerID == playerID {
			vp += 5
		}
	}
	return vp
}
