package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

// TriggerType represents different trigger conditions
type TriggerType string

const (
	TriggerOceanPlaced           TriggerType = "ocean-placed"
	TriggerGlobalParameterRaised TriggerType = "global-parameter-raised"
	TriggerCityPlaced            TriggerType = "city-placed"
	TriggerGreeneryPlaced        TriggerType = "greenery-placed"
	TriggerTilePlaced            TriggerType = "tile-placed"
	TriggerCardPlayed            TriggerType = "card-played"
	TriggerStandardProjectPlayed TriggerType = "standard-project-played"
	TriggerTagPlayed             TriggerType = "tag-played"
	TriggerProductionIncreased   TriggerType = "production-increased"
	TriggerPlacementBonusGained  TriggerType = "placement-bonus-gained"
	TriggerAlwaysActive          TriggerType = "always-active"
	TriggerCardHandUpdated       TriggerType = "card-hand-updated"
	TriggerPlayerEffectsChanged  TriggerType = "player-effects-changed"
)

// ResourceTriggerType represents trigger types for resource exchanges
type ResourceTriggerType string

const (
	ResourceTriggerManual                     ResourceTriggerType = "manual"
	ResourceTriggerAuto                       ResourceTriggerType = "auto"
	ResourceTriggerAutoCorporationFirstAction ResourceTriggerType = "auto-corporation-first-action"
	ResourceTriggerAutoCorporationStart       ResourceTriggerType = "auto-corporation-start"
)

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      ResourceTriggerType       `json:"type"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty"`
}

// MinMaxValue represents a min/max value constraint
type MinMaxValue struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                   TriggerType                         `json:"type"`
	Location               *CardApplyLocation                  `json:"location,omitempty"`
	Target                 *TargetType                         `json:"target,omitempty"`
	RequiredOriginalCost   *MinMaxValue                        `json:"requiredOriginalCost,omitempty"`
	RequiredResourceChange map[shared.ResourceType]MinMaxValue `json:"requiredResourceChange,omitempty"`
}
