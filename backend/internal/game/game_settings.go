package game

import (
	"slices"

	"terraforming-mars-backend/internal/game/global_parameters"
)

// GameSettings contains configurable game parameters (all optional)
type GameSettings struct {
	MaxPlayers       int      // Default: 10
	Temperature      *int     // Default: -30°C
	Oxygen           *int     // Default: 0%
	Oceans           *int     // Default: 0
	Venus            *int     // Default: 0%
	VenusNextEnabled bool     // Default: false
	DevelopmentMode  bool     // Default: false
	DemoGame         bool     // Default: false - enables lobby corp/card selection
	CardPacks        []string // Default: ["base-game"]
	ClaudeAPIKey     string   // API key for Claude bot invocations (secret, never exposed in DTOs)
	ClaudeModel      string   // Claude model to use for bots (default: "sonnet")
}

// Card pack constants
const (
	PackBaseGame = "base-game"  // Tested simple cards only
	PackFuture   = "future"     // Untested/complex cards for future implementation
	PackPrelude  = "prelude"    // Prelude expansion cards
	PackVenus    = "venus-next" // Venus Next expansion cards
)

// HasPrelude returns true if the prelude card pack is enabled
func (s GameSettings) HasPrelude() bool {
	return slices.Contains(s.CardPacks, PackPrelude)
}

// HasVenus returns true if the Venus expansion is enabled
func (s GameSettings) HasVenus() bool {
	return s.VenusNextEnabled
}

// Default values for game settings
const (
	DefaultMaxPlayers  = 10
	DefaultTemperature = global_parameters.MinTemperature // -30°C
	DefaultOxygen      = global_parameters.MinOxygen      // 0%
	DefaultOceans      = global_parameters.MinOceans      // 0
	DefaultVenus       = global_parameters.MinVenus       // 0%
)

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame, PackPrelude}
}
