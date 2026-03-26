package player

import (
	"terraforming-mars-backend/internal/game/shared"
	"time"
)

// EntityState holds calculated state for any entity (card, action, project).
// This generic structure eliminates redundant boolean flags and works across all entity types.
// The single source of truth is the Errors slice - availability is computed from it.
type EntityState struct {
	Errors         []StateError
	Warnings       []StateWarning
	Cost           map[string]int
	Metadata       map[string]any
	ComputedValues []ComputedBehaviorValue
	LastCalculated time.Time
}

// ComputedBehaviorValue holds pre-computed output values for a specific behavior.
// Target uses the format "behaviors::N" where N is the behavior index.
type ComputedBehaviorValue struct {
	Target  string
	Outputs []shared.CalculatedOutput
}

// Available returns true if there are no errors (computed, not stored).
// This prevents contradictory state between availability flags and error lists.
func (e EntityState) Available() bool {
	return len(e.Errors) == 0
}

// StateError represents a specific reason why an entity is unavailable.
// Errors are categorized for UI filtering and display.
type StateError struct {
	Code     StateErrorCode
	Category StateErrorCategory
	Message  string
}

// StateWarning represents a non-blocking warning about an action.
// Warnings inform the player of potential issues without preventing the action.
type StateWarning struct {
	Code    StateWarningCode
	Message string
}
