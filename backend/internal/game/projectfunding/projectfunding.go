package projectfunding

import "terraforming-mars-backend/internal/game/shared"

// ProjectDefinition is the static template loaded from JSON
type ProjectDefinition struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Seats            []SeatDefinition `json:"seats"`
	RewardTiers      []RewardTier     `json:"rewardTiers"`
	CompletionEffect CompletionEffect `json:"completionEffect"`
	Style            shared.Style     `json:"style"`
}

// SeatDefinition represents one purchasable seat in a project
type SeatDefinition struct {
	Cost               int                 `json:"cost"`
	PaymentSubstitutes []PaymentSubstitute `json:"paymentSubstitutes,omitempty"`
}

// PaymentSubstitute allows a resource to be used instead of credits at a conversion rate
type PaymentSubstitute struct {
	ResourceType   string `json:"resourceType"`
	ConversionRate int    `json:"conversionRate"`
}

// RewardTier defines rewards granted to a funder based on seats owned when project completes
type RewardTier struct {
	SeatsOwned int      `json:"seatsOwned"`
	Rewards    []Output `json:"rewards"`
}

// CompletionEffect defines rewards granted to ALL players when a project completes
type CompletionEffect struct {
	Description   string         `json:"description"`
	Rewards       []Output       `json:"rewards"`
	GlobalEffects []GlobalOutput `json:"globalEffects,omitempty"`
}

// GlobalOutput represents a one-time game-wide effect on project completion
type GlobalOutput struct {
	Type   string `json:"type"`             // "temperature", "oxygen", "freeze-turn-order", "production-choice", "card-draw"
	Amount int    `json:"amount,omitempty"` // used for card-draw (N cards)
}

// Output represents a resource gain (type + amount)
type Output struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// FindBestTier returns the highest reward tier the player qualifies for based on seats owned.
func FindBestTier(tiers []RewardTier, seatsOwned int) *RewardTier {
	var best *RewardTier
	for i := range tiers {
		if tiers[i].SeatsOwned <= seatsOwned {
			if best == nil || tiers[i].SeatsOwned > best.SeatsOwned {
				best = &tiers[i]
			}
		}
	}
	return best
}

// ProjectState is the runtime mutable state per project in a game
type ProjectState struct {
	DefinitionID string
	SeatOwners   []string // PlayerIDs in purchase order (same player can appear multiple times)
	IsCompleted  bool
}

// SeatCostBaselinePlayers is the player count for which JSON seat costs and seat counts are calibrated.
const SeatCostBaselinePlayers = 4

// ScaledSeatCost returns the seat cost adjusted for the number of players.
// JSON costs are calibrated for SeatCostBaselinePlayers; this scales linearly from there.
func ScaledSeatCost(baseCost, playerCount int) int {
	if baseCost <= 0 {
		return baseCost
	}
	scaled := baseCost * playerCount / SeatCostBaselinePlayers
	if scaled < 1 {
		return 1
	}
	return scaled
}

// ScaledSeatCount returns the seat count adjusted for the number of players.
// JSON counts are calibrated for SeatCostBaselinePlayers; rounds to nearest, floored at 1.
func ScaledSeatCount(baseCount, playerCount int) int {
	if baseCount <= 0 {
		return baseCount
	}
	n := (baseCount*playerCount + SeatCostBaselinePlayers/2) / SeatCostBaselinePlayers
	if n < 1 {
		return 1
	}
	return n
}

// ScaledSeats returns seat definitions scaled to the given player count.
// Seat costs are scaled via ScaledSeatCost. The list is truncated when shrinking
// and extended by repeating the final JSON entry when growing.
func ScaledSeats(seats []SeatDefinition, playerCount int) []SeatDefinition {
	if len(seats) == 0 {
		return nil
	}
	n := ScaledSeatCount(len(seats), playerCount)
	out := make([]SeatDefinition, n)
	for i := 0; i < n; i++ {
		idx := i
		if idx >= len(seats) {
			idx = len(seats) - 1
		}
		src := seats[idx]
		out[i] = SeatDefinition{
			Cost:               ScaledSeatCost(src.Cost, playerCount),
			PaymentSubstitutes: src.PaymentSubstitutes,
		}
	}
	return out
}

// ScaledRewardTiers returns reward tiers with seatsOwned thresholds scaled by player count.
// Thresholds are rounded to nearest and floored at 1.
func ScaledRewardTiers(tiers []RewardTier, playerCount int) []RewardTier {
	if len(tiers) == 0 {
		return nil
	}
	out := make([]RewardTier, len(tiers))
	for i, t := range tiers {
		scaled := (t.SeatsOwned*playerCount + SeatCostBaselinePlayers/2) / SeatCostBaselinePlayers
		if scaled < 1 {
			scaled = 1
		}
		out[i] = RewardTier{
			SeatsOwned: scaled,
			Rewards:    t.Rewards,
		}
	}
	return out
}
