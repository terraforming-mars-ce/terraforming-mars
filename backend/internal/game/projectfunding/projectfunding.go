package projectfunding

// ProjectDefinition is the static template loaded from JSON
type ProjectDefinition struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Seats            []SeatDefinition `json:"seats"`
	RewardTiers      []RewardTier     `json:"rewardTiers"`
	CompletionEffect CompletionEffect `json:"completionEffect"`
	Style            Style            `json:"style"`
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
	Description string   `json:"description"`
	Rewards     []Output `json:"rewards"`
}

// Output represents a resource gain (type + amount)
type Output struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// Style provides visual hints for the frontend
type Style struct {
	Color string `json:"color"`
	Icon  string `json:"icon"`
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
