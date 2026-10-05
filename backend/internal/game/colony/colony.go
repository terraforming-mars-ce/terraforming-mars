package colony

import "terraforming-mars-backend/internal/game/shared"

// ColonyDefinition is the static template loaded from JSON.
type ColonyDefinition struct {
	ActivationResource string       `json:"activationResource,omitempty"`
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Location           string       `json:"location"`
	Steps              []ColonyStep `json:"steps"`
	ColonyBonus        []Output     `json:"colonyBonus"`
	Colonies           []ColonySlot `json:"colonies"`
	Style              shared.Style `json:"style"`
}

// ColonyStep represents one position on the trade track
type ColonyStep struct {
	Outputs []Output `json:"outputs"`
}

// Output represents a resource gain (type + amount)
type Output struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// ColonySlot represents a colony placement slot with its placement reward
type ColonySlot struct {
	Reward []Output `json:"reward"`
}

// ColonyState is the runtime mutable state per colony in a game.
type ColonyState struct {
	AwaitingResource string
	DefinitionID     string
	MarkerPosition   int
	PlayerColonies   []string // PlayerIDs with colonies (max len(Colonies) from definition)
	TradedThisGen    bool
	TraderID         string // PlayerID who traded here this gen
}

// TradeFleet tracks permanent capacity and use during the current generation.
type TradeFleet struct {
	Capacity int
	Used     int
}

// Available returns the number of fleets that can still trade.
func (f TradeFleet) Available() int { return max(0, f.Capacity-f.Used) }

// NewTileState initializes the trade marker and resource activation requirement.
func NewTileState(definition ColonyDefinition, resourceInPlay bool) *ColonyState {
	state := &ColonyState{DefinitionID: definition.ID, MarkerPosition: 1, PlayerColonies: []string{}}
	if definition.ActivationResource != "" && !resourceInPlay {
		state.AwaitingResource = definition.ActivationResource
		state.MarkerPosition = -1
	}
	return state
}
