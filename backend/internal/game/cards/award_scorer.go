package cards

import (
	"sort"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Award VP constants
const (
	AwardFirstPlaceVP  = 5 // VP for first place in an award
	AwardSecondPlaceVP = 2 // VP for second place in an award
)

// AwardPlacement represents a player's placement in an award
type AwardPlacement struct {
	PlayerID  string
	Score     int
	Placement int // 1 = first place (5 VP), 2 = second place (2 VP), 0 = no placement
}

// CalculateAwardScore calculates a player's score for a specific award
func CalculateAwardScore(
	awardType shared.AwardType,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	switch awardType {
	case shared.AwardLandlord:
		// Most cities + greeneries on Mars (oceans don't count)
		cityType := shared.ResourceCityTile
		greeneryType := shared.ResourceGreeneryTile
		return CountPlayerTiles(p.ID(), b, &cityType) + CountPlayerTiles(p.ID(), b, &greeneryType)
	case shared.AwardBanker:
		// Highest MC production
		production := p.Resources().Production()
		return production.Credits
	case shared.AwardScientist:
		// Most science tags in play
		return CountPlayerTagsByType(p, cardRegistry, shared.TagScience)
	case shared.AwardThermalist:
		// Most heat resources
		resources := p.Resources().Get()
		return resources.Heat
	case shared.AwardMiner:
		// Most steel + titanium resources
		resources := p.Resources().Get()
		return resources.Steel + resources.Titanium
	default:
		return 0
	}
}

// ScoreAward calculates placements for all players for an award
// Returns a slice of AwardPlacement sorted by placement (1st, 2nd, then others)
func ScoreAward(
	awardType shared.AwardType,
	players []*player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) []AwardPlacement {
	placements := make([]AwardPlacement, len(players))
	for i, p := range players {
		placements[i] = AwardPlacement{
			PlayerID: p.ID(),
			Score:    CalculateAwardScore(awardType, p, b, cardRegistry),
		}
	}

	sort.Slice(placements, func(i, j int) bool {
		return placements[i].Score > placements[j].Score
	})

	if len(placements) == 0 {
		return placements
	}

	firstPlaceScore := placements[0].Score
	for i := range placements {
		if placements[i].Score == firstPlaceScore {
			placements[i].Placement = 1
		} else {
			break
		}
	}

	firstPlaceCount := 0
	for _, p := range placements {
		if p.Placement == 1 {
			firstPlaceCount++
		}
	}

	if firstPlaceCount < len(placements) {
		var secondPlaceScore int
		foundSecond := false
		for _, p := range placements {
			if p.Placement != 1 {
				secondPlaceScore = p.Score
				foundSecond = true
				break
			}
		}

		if foundSecond {
			for i := range placements {
				if placements[i].Placement == 0 && placements[i].Score == secondPlaceScore {
					placements[i].Placement = 2
				}
			}
		}
	}

	return placements
}

// GetAwardVP returns the VP for a specific placement
func GetAwardVP(placement int) int {
	switch placement {
	case 1:
		return AwardFirstPlaceVP
	case 2:
		return AwardSecondPlaceVP
	default:
		return 0
	}
}
