package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Milestone VP constant
const MilestoneClaimedVP = 5 // VP for each claimed milestone

// CardVPConditionDetail represents the detailed calculation of a single VP condition
type CardVPConditionDetail struct {
	ConditionType  string `json:"conditionType"`  // "fixed", "per", "once"
	Amount         int    `json:"amount"`         // VP amount per trigger or fixed amount
	Count          int    `json:"count"`          // Items counted (for "per" conditions)
	MaxTrigger     *int   `json:"maxTrigger"`     // Max triggers limit (if any)
	ActualTriggers int    `json:"actualTriggers"` // Actual triggers after applying max
	TotalVP        int    `json:"totalVP"`        // Final VP from this condition
	Explanation    string `json:"explanation"`    // Human-readable breakdown
}

// CardVPDetail represents VP calculation for a single card
type CardVPDetail struct {
	CardID     string                  `json:"cardId"`
	CardName   string                  `json:"cardName"`
	Conditions []CardVPConditionDetail `json:"conditions"`
	TotalVP    int                     `json:"totalVP"`
}

// GreeneryVPDetail represents VP from a single greenery tile
type GreeneryVPDetail struct {
	Coordinate string `json:"coordinate"` // Format: "q,r,s"
	VP         int    `json:"vp"`         // Always 1 per greenery
}

// CityVPDetail represents VP from a single city tile and its adjacent greeneries
type CityVPDetail struct {
	CityCoordinate     string   `json:"cityCoordinate"`     // Format: "q,r,s"
	AdjacentGreeneries []string `json:"adjacentGreeneries"` // Coordinates of adjacent greenery tiles
	VP                 int      `json:"vp"`                 // Number of adjacent greeneries
}

// VPBreakdown contains the detailed breakdown of a player's victory points
type VPBreakdown struct {
	TerraformRating   int                `json:"terraformRating"`
	CardVP            int                `json:"cardVP"`
	CardVPDetails     []CardVPDetail     `json:"cardVPDetails"` // Per-card VP breakdown
	MilestoneVP       int                `json:"milestoneVP"`
	AwardVP           int                `json:"awardVP"`
	GreeneryVP        int                `json:"greeneryVP"`
	GreeneryVPDetails []GreeneryVPDetail `json:"greeneryVPDetails"` // Per-greenery VP breakdown
	CityVP            int                `json:"cityVP"`
	CityVPDetails     []CityVPDetail     `json:"cityVPDetails"` // Per-city VP breakdown with adjacencies
	TotalVP           int                `json:"totalVP"`
}

// MilestonesInterface defines the interface for accessing milestones
type MilestonesInterface interface {
	GetClaimedByPlayer(playerID string) []ClaimedMilestoneInfo
}

// ClaimedMilestoneInfo represents info about a claimed milestone
type ClaimedMilestoneInfo struct {
	Type     string
	PlayerID string
}

// AwardsInterface defines the interface for accessing funded awards
type AwardsInterface interface {
	FundedAwards() []FundedAwardInfo
}

// FundedAwardInfo represents info about a funded award
type FundedAwardInfo struct {
	Type string
}

// CalculatePlayerVP computes the total VP for a player with detailed breakdown
func CalculatePlayerVP(
	p *player.Player,
	b *board.Board,
	claimedMilestones []ClaimedMilestoneInfo,
	fundedAwards []FundedAwardInfo,
	allPlayers []*player.Player,
	cardRegistry CardRegistryInterface,
) VPBreakdown {
	breakdown := VPBreakdown{}

	// 1. Terraform Rating
	breakdown.TerraformRating = p.Resources().TerraformRating()

	// 2. Card VP (with detailed breakdown)
	cardVPDetails := calculateCardVPDetailed(p, b, cardRegistry)
	breakdown.CardVPDetails = cardVPDetails
	breakdown.CardVP = 0
	for _, detail := range cardVPDetails {
		breakdown.CardVP += detail.TotalVP
	}

	// 3. Milestone VP (5 VP per claimed milestone)
	breakdown.MilestoneVP = calculateMilestoneVP(p.ID(), claimedMilestones)

	// 4. Award VP
	breakdown.AwardVP = calculateAwardVP(p.ID(), fundedAwards, allPlayers, b, cardRegistry)

	// 5. Greenery VP (1 VP per greenery tile owned) with detailed breakdown
	greeneryDetails := calculateGreeneryVPDetailed(p.ID(), b)
	breakdown.GreeneryVPDetails = greeneryDetails
	breakdown.GreeneryVP = 0
	for _, detail := range greeneryDetails {
		breakdown.GreeneryVP += detail.VP
	}

	// 6. City VP (1 VP per adjacent greenery to owned cities) with detailed breakdown
	cityDetails := calculateCityVPDetailed(p.ID(), b)
	breakdown.CityVPDetails = cityDetails
	breakdown.CityVP = 0
	for _, detail := range cityDetails {
		breakdown.CityVP += detail.VP
	}

	breakdown.TotalVP = breakdown.TerraformRating +
		breakdown.CardVP +
		breakdown.MilestoneVP +
		breakdown.AwardVP +
		breakdown.GreeneryVP +
		breakdown.CityVP

	return breakdown
}

// calculateCardVPDetailed calculates VP from all played cards with detailed breakdown
func calculateCardVPDetailed(p *player.Player, b *board.Board, cardRegistry CardRegistryInterface) []CardVPDetail {
	var details []CardVPDetail
	playedCardIDs := p.PlayedCards().Cards()

	for _, cardID := range playedCardIDs {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue // Skip cards not in registry
		}

		if len(card.VPConditions) == 0 {
			continue // Skip cards with no VP
		}

		detail := CardVPDetail{
			CardID:   card.ID,
			CardName: card.Name,
			TotalVP:  0,
		}

		for _, vpCond := range card.VPConditions {
			condDetail := evaluateVPConditionDetailed(vpCond, p, b, card, cardRegistry)
			detail.Conditions = append(detail.Conditions, condDetail)
			detail.TotalVP += condDetail.TotalVP
		}

		if detail.TotalVP > 0 || len(detail.Conditions) > 0 {
			details = append(details, detail)
		}
	}

	return details
}

// evaluateVPConditionDetailed evaluates a single VP condition and returns detailed breakdown
func evaluateVPConditionDetailed(
	vpCond VictoryPointCondition,
	p *player.Player,
	b *board.Board,
	card *Card,
	cardRegistry CardRegistryInterface,
) CardVPConditionDetail {
	detail := CardVPConditionDetail{
		ConditionType: string(vpCond.Condition),
		Amount:        vpCond.Amount,
	}

	switch vpCond.Condition {
	case VPConditionFixed:
		detail.ActualTriggers = 1
		detail.TotalVP = vpCond.Amount
		detail.Explanation = fmt.Sprintf("%d VP", vpCond.Amount)

	case VPConditionPer:
		if vpCond.Per == nil {
			detail.Explanation = "Invalid condition"
			return detail
		}

		count := countPerCondition(vpCond.Per, p, b, card, cardRegistry)
		detail.Count = count

		if vpCond.Per.Amount > 0 {
			triggers := count / vpCond.Per.Amount
			detail.MaxTrigger = vpCond.MaxTrigger

			if vpCond.MaxTrigger != nil && *vpCond.MaxTrigger >= 0 && triggers > *vpCond.MaxTrigger {
				triggers = *vpCond.MaxTrigger
			}
			detail.ActualTriggers = triggers
			detail.TotalVP = vpCond.Amount * triggers

			// Build human-readable explanation
			countType := getPerConditionTypeName(vpCond.Per)
			if vpCond.Per.Amount == 1 {
				detail.Explanation = fmt.Sprintf("%d VP (%d %s)", detail.TotalVP, count, countType)
			} else {
				detail.Explanation = fmt.Sprintf("%d VP (%d per %d %s, %d found)", detail.TotalVP, vpCond.Amount, vpCond.Per.Amount, countType, count)
			}
		}

	case VPConditionOnce:
		detail.ActualTriggers = 1
		detail.TotalVP = vpCond.Amount
		detail.Explanation = fmt.Sprintf("%d VP (one-time)", vpCond.Amount)

	default:
		detail.Explanation = "Unknown condition"
	}

	return detail
}

// getPerConditionTypeName returns a human-readable name for the per condition type
func getPerConditionTypeName(per *PerCondition) string {
	if per.Target != nil && *per.Target == TargetSelfCard {
		return "on this card"
	}
	if per.Tag != nil {
		return string(*per.Tag) + " tags"
	}
	switch per.Type {
	case shared.ResourceOceanTile:
		return "ocean tiles"
	case shared.ResourceCityTile:
		return "city tiles"
	case shared.ResourceGreeneryTile:
		return "greenery tiles"
	default:
		return string(per.Type)
	}
}

// countPerCondition counts the number of items matching a per condition
func countPerCondition(
	per *PerCondition,
	p *player.Player,
	b *board.Board,
	card *Card,
	cardRegistry CardRegistryInterface,
) int {
	// Handle resource storage on card (e.g., animals on this card)
	if per.Target != nil && *per.Target == TargetSelfCard {
		storage := p.Resources().GetCardStorage(card.ID)
		return storage
	}

	// Handle adjacency-based counting (e.g., World Tree: 1 VP per adjacent forest)
	if per.AdjacentToTileType != nil {
		return countAdjacentTilesOfType(p.ID(), b, per.Type, *per.AdjacentToTileType)
	}

	cityTileType := shared.ResourceCityTile
	greeneryTileType := shared.ResourceGreeneryTile

	switch per.Type {
	case shared.ResourceOceanTile:
		return CountAllTilesOfType(b, shared.ResourceOceanTile)
	case shared.ResourceCityTile:
		if per.Target != nil && *per.Target == TargetSelfPlayer {
			return CountPlayerTiles(p.ID(), b, &cityTileType)
		}
		return CountAllTilesOfType(b, shared.ResourceCityTile)
	case shared.ResourceGreeneryTile:
		if per.Target != nil && *per.Target == TargetSelfPlayer {
			return CountPlayerTiles(p.ID(), b, &greeneryTileType)
		}
		return CountAllTilesOfType(b, shared.ResourceGreeneryTile)
	default:
		// Check if it's a tag type
		if per.Tag != nil {
			return CountPlayerTagsByType(p, cardRegistry, *per.Tag)
		}
		// Default: try to count as a tag
		return CountPlayerTagsByType(p, cardRegistry, shared.CardTag(per.Type))
	}
}

// calculateMilestoneVP calculates VP from claimed milestones
func calculateMilestoneVP(playerID string, claimedMilestones []ClaimedMilestoneInfo) int {
	vp := 0
	for _, milestone := range claimedMilestones {
		if milestone.PlayerID == playerID {
			vp += MilestoneClaimedVP
		}
	}
	return vp
}

// calculateAwardVP calculates VP from awards
func calculateAwardVP(
	playerID string,
	fundedAwards []FundedAwardInfo,
	allPlayers []*player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	totalVP := 0

	for _, award := range fundedAwards {
		placements := ScoreAward(shared.AwardType(award.Type), allPlayers, b, cardRegistry)
		for _, placement := range placements {
			if placement.PlayerID == playerID {
				totalVP += GetAwardVP(placement.Placement)
				break
			}
		}
	}

	return totalVP
}

// calculateGreeneryVPDetailed calculates VP from greenery tiles with coordinate details
func calculateGreeneryVPDetailed(playerID string, b *board.Board) []GreeneryVPDetail {
	var details []GreeneryVPDetail
	tiles := b.Tiles()

	for _, tile := range tiles {
		if tile.OwnerID == nil || *tile.OwnerID != playerID {
			continue
		}
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != shared.ResourceGreeneryTile {
			continue
		}

		coord := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
		details = append(details, GreeneryVPDetail{
			Coordinate: coord,
			VP:         1,
		})
	}

	return details
}

// calculateCityVPDetailed calculates VP from city tiles with adjacent greenery coordinates
func calculateCityVPDetailed(playerID string, b *board.Board) []CityVPDetail {
	var details []CityVPDetail
	tiles := b.Tiles()

	// Find all cities owned by the player
	for _, tile := range tiles {
		if tile.OwnerID == nil || *tile.OwnerID != playerID {
			continue
		}
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != shared.ResourceCityTile {
			continue
		}

		cityCoord := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
		adjacentGreeneryCoords := getAdjacentGreeneryCoordinates(tile.Coordinates, tiles)

		details = append(details, CityVPDetail{
			CityCoordinate:     cityCoord,
			AdjacentGreeneries: adjacentGreeneryCoords,
			VP:                 len(adjacentGreeneryCoords),
		})
	}

	return details
}

// getAdjacentGreeneryCoordinates returns coordinates of greenery tiles adjacent to a position
func getAdjacentGreeneryCoordinates(coords shared.HexPosition, tiles []board.Tile) []string {
	var greeneryCoords []string
	neighbors := coords.GetNeighbors()

	for _, tile := range tiles {
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != shared.ResourceGreeneryTile {
			continue
		}
		for _, neighbor := range neighbors {
			if tile.Coordinates == neighbor {
				coord := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
				greeneryCoords = append(greeneryCoords, coord)
				break
			}
		}
	}

	return greeneryCoords
}

// isForestTile returns true if the tile type counts as a forest (greenery or world-tree)
func isForestTile(tileType shared.ResourceType) bool {
	return tileType == shared.ResourceGreeneryTile || tileType == shared.ResourceWorldTreeTile
}

// countAdjacentTilesOfType counts unique tiles matching countType that are adjacent
// to tiles of adjacentToType owned by playerID. Uses isForestTile for greenery matching.
func countAdjacentTilesOfType(playerID string, b *board.Board, countType shared.ResourceType, adjacentToType shared.ResourceType) int {
	tiles := b.Tiles()
	counted := make(map[shared.HexPosition]bool)

	for _, tile := range tiles {
		if tile.OwnerID == nil || *tile.OwnerID != playerID {
			continue
		}
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != adjacentToType {
			continue
		}

		neighbors := tile.Coordinates.GetNeighbors()
		for _, neighborTile := range tiles {
			if neighborTile.OccupiedBy == nil || counted[neighborTile.Coordinates] {
				continue
			}

			var matches bool
			if countType == shared.ResourceGreeneryTile {
				matches = isForestTile(neighborTile.OccupiedBy.Type)
			} else {
				matches = neighborTile.OccupiedBy.Type == countType
			}
			if !matches {
				continue
			}

			for _, neighbor := range neighbors {
				if neighborTile.Coordinates == neighbor {
					counted[neighborTile.Coordinates] = true
					break
				}
			}
		}
	}

	return len(counted)
}
