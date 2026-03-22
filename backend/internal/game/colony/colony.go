package colony

import "terraforming-mars-backend/internal/game/shared"

// ColonyTileDefinition is the static template loaded from JSON
type ColonyTileDefinition struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Location    string       `json:"location"`
	Steps       []ColonyStep `json:"steps"`
	ColonyBonus []Output     `json:"colonyBonus"`
	Colonies    []ColonySlot `json:"colonies"`
	Style       shared.Style `json:"style"`
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

// TileState is the runtime mutable state per colony tile in a game
type TileState struct {
	DefinitionID   string
	MarkerPosition int
	PlayerColonies []string // PlayerIDs with colonies (max len(Colonies) from definition)
	TradedThisGen  bool
	TraderID       string // PlayerID who traded here this gen
}
