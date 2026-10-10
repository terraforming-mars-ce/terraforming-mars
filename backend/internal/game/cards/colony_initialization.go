package cards

import (
	"openmars/internal/game"
	"openmars/internal/game/colony"
)

// InitializeColonyTile creates a tile with its normal activation and track state.
func InitializeColonyTile(g *game.Game, definition colony.ColonyDefinition, registry CardRegistryInterface) *colony.ColonyState {
	state := colony.NewTileState(definition, false)
	if state.AwaitingResource == "" {
		return state
	}
	for _, player := range g.GetAllPlayers() {
		ids := append(player.PlayedCards().Cards(), player.CorporationID())
		for _, id := range ids {
			card, err := registry.GetByID(id)
			if err == nil && card.ResourceStorage != nil && string(card.ResourceStorage.Type) == state.AwaitingResource {
				state.AwaitingResource = ""
				state.MarkerPosition = 1
				return state
			}
		}
	}
	state.MarkerPosition = -1
	return state
}
