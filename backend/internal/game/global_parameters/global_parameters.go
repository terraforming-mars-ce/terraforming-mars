package global_parameters

import (
	"context"
	"sync"
	"terraforming-mars-backend/internal/events"
)

// Constants for terraforming limits
const (
	MinTemperature = -30
	MaxTemperature = 8
	MinOxygen      = 0
	MaxOxygen      = 14
	MinOceans      = 0
	MaxOceans      = 9
	MinVenus       = 0
	MaxVenus       = 30
)

// GlobalParameters represents the terraforming progress with encapsulated state
type GlobalParameters struct {
	mu          sync.RWMutex
	gameID      string
	temperature int // Range: -30 to +8°C
	oxygen      int // Range: 0-14%
	oceans      int // Range: 0-9
	venus       int // Range: 0-30%
	eventBus    *events.EventBusImpl
}

// NewGlobalParameters creates a new GlobalParameters instance with default starting values
func NewGlobalParameters(gameID string, eventBus *events.EventBusImpl) *GlobalParameters {
	return &GlobalParameters{
		gameID:      gameID,
		temperature: MinTemperature,
		oxygen:      MinOxygen,
		oceans:      MinOceans,
		venus:       MinVenus,
		eventBus:    eventBus,
	}
}

// NewGlobalParametersWithValues creates a new GlobalParameters instance with custom starting values
func NewGlobalParametersWithValues(gameID string, temperature, oxygen, oceans, venus int, eventBus *events.EventBusImpl) *GlobalParameters {
	return &GlobalParameters{
		gameID:      gameID,
		temperature: temperature,
		oxygen:      oxygen,
		oceans:      oceans,
		venus:       venus,
		eventBus:    eventBus,
	}
}

// Temperature returns the current temperature
func (gp *GlobalParameters) Temperature() int {
	gp.mu.RLock()
	defer gp.mu.RUnlock()
	return gp.temperature
}

// Oxygen returns the current oxygen level
func (gp *GlobalParameters) Oxygen() int {
	gp.mu.RLock()
	defer gp.mu.RUnlock()
	return gp.oxygen
}

// Oceans returns the current ocean count
func (gp *GlobalParameters) Oceans() int {
	gp.mu.RLock()
	defer gp.mu.RUnlock()
	return gp.oceans
}

// Venus returns the current venus level
func (gp *GlobalParameters) Venus() int {
	gp.mu.RLock()
	defer gp.mu.RUnlock()
	return gp.venus
}

// IsMaxed returns true if all global parameters have reached their maximum values
func (gp *GlobalParameters) IsMaxed() bool {
	gp.mu.RLock()
	defer gp.mu.RUnlock()
	return gp.temperature >= MaxTemperature &&
		gp.oxygen >= MaxOxygen &&
		gp.oceans >= MaxOceans
}

// IncreaseTemperature raises the temperature by the specified number of steps
// Each step is 2 degrees. Returns the actual number of steps raised (may be less if limit reached)
// Publishes TemperatureChangedEvent after state change
func (gp *GlobalParameters) IncreaseTemperature(ctx context.Context, steps int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldTemp, newTemp int
	var actualSteps int

	// Critical: Capture values while holding lock, publish AFTER releasing
	gp.mu.Lock()
	oldTemp = gp.temperature
	newTemp = gp.temperature + (steps * 2) // Each step is 2 degrees
	if newTemp > MaxTemperature {
		newTemp = MaxTemperature
	}
	gp.temperature = newTemp
	actualSteps = (newTemp - oldTemp) / 2
	gp.mu.Unlock()

	// Publish event AFTER releasing lock to avoid deadlocks
	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldTemp != newTemp {
		events.Publish(gp.eventBus, events.TemperatureChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldTemp,
			NewValue: newTemp,
		})
	}

	return actualSteps, nil
}

// IncreaseOxygen raises the oxygen by the specified number of steps
// Returns the actual number of steps raised (may be less if limit reached)
// Publishes OxygenChangedEvent after state change
func (gp *GlobalParameters) IncreaseOxygen(ctx context.Context, steps int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldOxygen, newOxygen int

	gp.mu.Lock()
	oldOxygen = gp.oxygen
	newOxygen = gp.oxygen + steps
	if newOxygen > MaxOxygen {
		newOxygen = MaxOxygen
	}
	gp.oxygen = newOxygen
	actualSteps := newOxygen - oldOxygen
	gp.mu.Unlock()

	// Publish event AFTER releasing lock
	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldOxygen != newOxygen {
		events.Publish(gp.eventBus, events.OxygenChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOxygen,
			NewValue: newOxygen,
		})
	}

	return actualSteps, nil
}

// PlaceOcean places an ocean tile (increments ocean count)
// Returns true if successful, false if limit reached
// Publishes OceansChangedEvent after state change
func (gp *GlobalParameters) PlaceOcean(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var oldOceans, newOceans int
	var success bool

	gp.mu.Lock()
	oldOceans = gp.oceans
	if gp.oceans >= MaxOceans {
		success = false
	} else {
		gp.oceans++
		success = true
	}
	newOceans = gp.oceans
	gp.mu.Unlock()

	// Publish event AFTER releasing lock
	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldOceans != newOceans {
		events.Publish(gp.eventBus, events.OceansChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOceans,
			NewValue: newOceans,
		})
	}

	return success, nil
}

// SetTemperature sets the temperature to a specific value (for admin/testing)
// Publishes TemperatureChangedEvent if value changed
func (gp *GlobalParameters) SetTemperature(ctx context.Context, newTemp int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldTemp int

	gp.mu.Lock()
	oldTemp = gp.temperature
	gp.temperature = newTemp
	gp.mu.Unlock()

	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldTemp != newTemp {
		events.Publish(gp.eventBus, events.TemperatureChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldTemp,
			NewValue: newTemp,
		})
	}

	return nil
}

// SetOxygen sets the oxygen to a specific value (for admin/testing)
// Publishes OxygenChangedEvent if value changed
func (gp *GlobalParameters) SetOxygen(ctx context.Context, newOxygen int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldOxygen int

	gp.mu.Lock()
	oldOxygen = gp.oxygen
	gp.oxygen = newOxygen
	gp.mu.Unlock()

	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldOxygen != newOxygen {
		events.Publish(gp.eventBus, events.OxygenChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOxygen,
			NewValue: newOxygen,
		})
	}

	return nil
}

// SetOceans sets the ocean count to a specific value (for admin/testing)
// Publishes OceansChangedEvent if value changed
func (gp *GlobalParameters) SetOceans(ctx context.Context, newOceans int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldOceans int

	gp.mu.Lock()
	oldOceans = gp.oceans
	gp.oceans = newOceans
	gp.mu.Unlock()

	// Automatic broadcasting handled by EventBus
	if gp.eventBus != nil && oldOceans != newOceans {
		events.Publish(gp.eventBus, events.OceansChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOceans,
			NewValue: newOceans,
		})
	}

	return nil
}

// IncreaseVenus raises the venus level by the specified number of steps
// Each step is 2%. Returns the actual number of steps raised (may be less if limit reached)
// Publishes VenusChangedEvent after state change
func (gp *GlobalParameters) IncreaseVenus(ctx context.Context, steps int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldVenus, newVenus int
	var actualSteps int

	gp.mu.Lock()
	oldVenus = gp.venus
	newVenus = gp.venus + (steps * 2)
	if newVenus > MaxVenus {
		newVenus = MaxVenus
	}
	gp.venus = newVenus
	actualSteps = (newVenus - oldVenus) / 2
	gp.mu.Unlock()

	if gp.eventBus != nil && oldVenus != newVenus {
		events.Publish(gp.eventBus, events.VenusChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldVenus,
			NewValue: newVenus,
		})
	}

	return actualSteps, nil
}

// SetVenus sets the venus level to a specific value (for admin/testing)
// Publishes VenusChangedEvent if value changed
func (gp *GlobalParameters) SetVenus(ctx context.Context, newVenus int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldVenus int

	gp.mu.Lock()
	oldVenus = gp.venus
	gp.venus = newVenus
	gp.mu.Unlock()

	if gp.eventBus != nil && oldVenus != newVenus {
		events.Publish(gp.eventBus, events.VenusChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldVenus,
			NewValue: newVenus,
		})
	}

	return nil
}
