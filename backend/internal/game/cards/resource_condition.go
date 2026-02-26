package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

// TargetType represents different targeting scopes
type TargetType string

const (
	TargetSelfPlayer   TargetType = "self-player"
	TargetSelfCard     TargetType = "self-card"
	TargetAnyCard      TargetType = "any-card"
	TargetAnyPlayer    TargetType = "any-player"
	TargetOtherPlayers TargetType = "other-players"
	TargetOpponent     TargetType = "opponent"
	TargetNone         TargetType = "none"
)

// TileRestrictions represents restrictions for tile placement
type TileRestrictions struct {
	BoardTags []string `json:"boardTags,omitempty"`
	Adjacency string   `json:"adjacency,omitempty"` // "none" = no adjacent occupied tiles
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	Type             shared.ResourceType `json:"type"`
	Amount           int                 `json:"amount"`
	Target           TargetType          `json:"target"`
	MaxTrigger       *int                `json:"maxTrigger,omitempty"`
	Per              *PerCondition       `json:"per,omitempty"`
	TileRestrictions *TileRestrictions   `json:"tileRestrictions,omitempty"`
}

// PerCondition represents what to count for conditional resource gains
type PerCondition struct {
	Type     shared.ResourceType `json:"type"`
	Amount   int                 `json:"amount"`
	Location *CardApplyLocation  `json:"location,omitempty"`
	Target   *TargetType         `json:"target,omitempty"`
	Tag      *shared.CardTag     `json:"tag,omitempty"`
}
