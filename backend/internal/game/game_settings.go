package game

import (
	"terraforming-mars-backend/internal/game/global_parameters"
)

// Default values for game settings
const (
	DefaultMaxPlayers  = 10
	DefaultTemperature = global_parameters.MinTemperature // -30°C
	DefaultOxygen      = global_parameters.MinOxygen      // 0%
	DefaultOceans      = global_parameters.MinOceans      // 0
	DefaultVenus       = global_parameters.MinVenus       // 0%
)
