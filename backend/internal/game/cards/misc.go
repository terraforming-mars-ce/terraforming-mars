package cards

import (
	"openmars/internal/game/shared"
)

// DiscountEffect represents cost reductions for playing cards
type DiscountEffect struct {
	Amount      int
	Tags        []shared.CardTag
	Description string
}

// ProductionEffects represents changes to resource production
type ProductionEffects struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// TerraformingActions represents tile placement actions
type TerraformingActions struct {
	CityPlacement     int
	OceanPlacement    int
	GreeneryPlacement int
}

// EffectContext provides context about a game event that triggered passive effects
type EffectContext struct {
	TriggeringPlayerID string
	TileCoordinate     *shared.HexPosition
	CardID             *string
	TagType            *shared.CardTag
	TileType           *shared.ResourceType
	ParameterChange    *int
}
