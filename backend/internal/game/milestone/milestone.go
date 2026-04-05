package milestone

import (
	"terraforming-mars-backend/internal/game/award"
	"terraforming-mars-backend/internal/game/shared"
)

// MilestoneDefinition is the static template loaded from JSON
type MilestoneDefinition struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Pack        string               `json:"pack"`
	ClaimCost   int                  `json:"claimCost"`
	Reward      []award.RewardOutput `json:"reward"`
	Requirement MilestoneRequirement `json:"requirement"`
	Style       shared.Style         `json:"style"`
}

const (
	RequirementKindCountable         = "countable"
	RequirementKindAllProductionsMin = "all-productions-min"
)

// MilestoneRequirement is a discriminated union for milestone requirements.
// The Kind field selects which sub-field is active.
type MilestoneRequirement struct {
	Kind              string                        `json:"kind"`
	Countable         *CountableRequirement         `json:"countable,omitempty"`
	AllProductionsMin *AllProductionsMinRequirement `json:"allProductionsMin,omitempty"`
}

// AllProductionsMinRequirement requires all 6 production types to be at or above a minimum.
type AllProductionsMinRequirement struct {
	Min int `json:"min"`
}

// CountableRequirement defines a countable threshold requirement.
// Uses PerCondition to determine what to count and MinMaxValue for the threshold.
type CountableRequirement struct {
	shared.PerCondition
	shared.MinMaxValue
}

// GetRewardVP returns the total VP from the reward outputs
func (d *MilestoneDefinition) GetRewardVP() int {
	vp := 0
	for _, o := range d.Reward {
		if o.Type == "vp" {
			vp += o.Amount
		}
	}
	return vp
}

// GetRequired returns the threshold value for the milestone requirement
func (d *MilestoneDefinition) GetRequired() int {
	switch d.Requirement.Kind {
	case RequirementKindCountable:
		if d.Requirement.Countable != nil {
			if d.Requirement.Countable.Min != nil {
				return *d.Requirement.Countable.Min
			}
			if d.Requirement.Countable.Max != nil {
				return *d.Requirement.Countable.Max
			}
		}
	case RequirementKindAllProductionsMin:
		return 6
	}
	return 0
}
