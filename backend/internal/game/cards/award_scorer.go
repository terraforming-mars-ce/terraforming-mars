package cards

import (
	"sort"

	"terraforming-mars-backend/internal/game/award"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
)

// AwardPlacement represents a player's placement in an award
type AwardPlacement struct {
	PlayerID  string
	Score     int
	Placement int // 1 = first place, 2 = second place, 0 = no placement
}

// CalculateAwardScore calculates a player's score for an award using its quantifier definition
func CalculateAwardScore(
	def *award.AwardDefinition,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	total := 0
	for _, q := range def.Quantifier {
		total += CountPerCondition(&q, "", p, b, cardRegistry, nil)
	}
	return total
}

// ScoreAward calculates placements for all players for an award
func ScoreAward(
	def *award.AwardDefinition,
	players []*player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) []AwardPlacement {
	placements := make([]AwardPlacement, len(players))
	for i, p := range players {
		placements[i] = AwardPlacement{
			PlayerID: p.ID(),
			Score:    CalculateAwardScore(def, p, b, cardRegistry),
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

// GetAwardVP returns the VP for a specific placement using the award definition
func GetAwardVP(def *award.AwardDefinition, placement int) int {
	return def.GetRewardVP(placement)
}
