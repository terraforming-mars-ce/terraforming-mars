package game

import (
	"terraforming-mars-backend/internal/game/global_parameters"
)

// GameSettings contains configurable game parameters (all optional)
type GameSettings struct {
	MaxPlayers      int      // Default: 5
	Temperature     *int     // Default: -30°C
	Oxygen          *int     // Default: 0%
	Oceans          *int     // Default: 0
	Venus           *int     // Default: 0%
	DevelopmentMode bool     // Default: false
	DemoGame        bool     // Default: false - enables lobby corp/card selection
	CardPacks       []string // Default: ["base-game"]
}

// Card pack constants
const (
	PackBaseGame = "base-game" // Tested simple cards only
	PackFuture   = "future"    // Untested/complex cards for future implementation
)

// Default values for game settings
const (
	DefaultMaxPlayers  = 5
	DefaultTemperature = global_parameters.MinTemperature // -30°C
	DefaultOxygen      = global_parameters.MinOxygen      // 0%
	DefaultOceans      = global_parameters.MinOceans      // 0
	DefaultVenus       = global_parameters.MinVenus       // 0%
)

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame}
}
