package global_parameters

import (
	"context"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/logger"
)

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

// GlobalParameters manages the global parameter fields of a game.
type GlobalParameters struct {
	ds       *datastore.DataStore
	gameID   string
	eventBus *events.EventBusImpl
}

func NewGlobalParameters(ds *datastore.DataStore, gameID string, eventBus *events.EventBusImpl) *GlobalParameters {
	return &GlobalParameters{ds: ds, gameID: gameID, eventBus: eventBus}
}

func (gp *GlobalParameters) update(fn func(s *datastore.GameState)) {
	if err := gp.ds.UpdateGame(gp.gameID, fn); err != nil {
		logger.Get().Warn("Failed to update game state", zap.String("game_id", gp.gameID), zap.Error(err))
	}
}

func (gp *GlobalParameters) read(fn func(s *datastore.GameState)) {
	if err := gp.ds.ReadGame(gp.gameID, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", gp.gameID), zap.Error(err))
	}
}

func (gp *GlobalParameters) Temperature() int {
	var v int
	gp.read(func(s *datastore.GameState) { v = s.Temperature })
	return v
}

func (gp *GlobalParameters) Oxygen() int {
	var v int
	gp.read(func(s *datastore.GameState) { v = s.Oxygen })
	return v
}

func (gp *GlobalParameters) Oceans() int {
	var v int
	gp.read(func(s *datastore.GameState) { v = s.Oceans })
	return v
}

func (gp *GlobalParameters) Venus() int {
	var v int
	gp.read(func(s *datastore.GameState) { v = s.Venus })
	return v
}

func (gp *GlobalParameters) GetMaxOceans() int {
	var v int
	gp.read(func(s *datastore.GameState) { v = s.MaxOceans })
	return v
}

// ReduceMaxOceans lowers the max oceans limit when non-ocean tiles occupy ocean spaces.
func (gp *GlobalParameters) ReduceMaxOceans(newMax int) {
	var oldMax, currentMax int
	gp.update(func(s *datastore.GameState) {
		oldMax = s.MaxOceans
		if newMax < s.MaxOceans {
			s.MaxOceans = newMax
		}
		currentMax = s.MaxOceans
	})

	if gp.eventBus != nil && oldMax != currentMax {
		var oceans int
		gp.read(func(s *datastore.GameState) { oceans = s.Oceans })
		events.Publish(gp.eventBus, events.OceansChangedEvent{
			GameID:   gp.gameID,
			OldValue: oceans,
			NewValue: oceans,
		})
	}
}

func (gp *GlobalParameters) IsMaxed() bool {
	var maxed bool
	gp.read(func(s *datastore.GameState) {
		maxed = s.Temperature >= MaxTemperature &&
			s.Oxygen >= MaxOxygen &&
			s.Oceans >= s.MaxOceans
	})
	return maxed
}

// IncreaseTemperature raises the temperature by the specified number of steps.
// Each step is 2 degrees. Returns the actual number of steps raised.
func (gp *GlobalParameters) IncreaseTemperature(ctx context.Context, steps int, playerID string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldTemp, newTemp int
	gp.update(func(s *datastore.GameState) {
		oldTemp = s.Temperature
		newTemp = min(s.Temperature+steps*2, MaxTemperature)
		s.Temperature = newTemp
	})
	actualSteps := (newTemp - oldTemp) / 2

	if gp.eventBus != nil && oldTemp != newTemp {
		events.Publish(gp.eventBus, events.TemperatureChangedEvent{
			GameID:    gp.gameID,
			OldValue:  oldTemp,
			NewValue:  newTemp,
			ChangedBy: playerID,
		})
	}

	return actualSteps, nil
}

// IncreaseOxygen raises the oxygen by the specified number of steps.
// Returns the actual number of steps raised.
func (gp *GlobalParameters) IncreaseOxygen(ctx context.Context, steps int, playerID string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldOxygen, newOxygen int
	gp.update(func(s *datastore.GameState) {
		oldOxygen = s.Oxygen
		newOxygen = min(s.Oxygen+steps, MaxOxygen)
		s.Oxygen = newOxygen
	})
	actualSteps := newOxygen - oldOxygen

	if gp.eventBus != nil && oldOxygen != newOxygen {
		events.Publish(gp.eventBus, events.OxygenChangedEvent{
			GameID:    gp.gameID,
			OldValue:  oldOxygen,
			NewValue:  newOxygen,
			ChangedBy: playerID,
		})
	}

	return actualSteps, nil
}

// PlaceOcean places an ocean tile (increments ocean count).
// Returns true if successful, false if limit reached.
func (gp *GlobalParameters) PlaceOcean(ctx context.Context, playerID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var oldOceans, newOceans int
	var placed bool
	gp.update(func(s *datastore.GameState) {
		oldOceans = s.Oceans
		if s.Oceans >= s.MaxOceans {
			newOceans = s.Oceans
			return
		}
		s.Oceans++
		newOceans = s.Oceans
		placed = true
	})

	if gp.eventBus != nil && placed {
		events.Publish(gp.eventBus, events.OceansChangedEvent{
			GameID:    gp.gameID,
			OldValue:  oldOceans,
			NewValue:  newOceans,
			ChangedBy: playerID,
		})
	}

	return placed, nil
}

func (gp *GlobalParameters) SetTemperature(ctx context.Context, newTemp int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldTemp int
	gp.update(func(s *datastore.GameState) {
		oldTemp = s.Temperature
		s.Temperature = newTemp
	})

	if gp.eventBus != nil && oldTemp != newTemp {
		events.Publish(gp.eventBus, events.TemperatureChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldTemp,
			NewValue: newTemp,
		})
	}

	return nil
}

func (gp *GlobalParameters) SetOxygen(ctx context.Context, newOxygen int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldOxygen int
	gp.update(func(s *datastore.GameState) {
		oldOxygen = s.Oxygen
		s.Oxygen = newOxygen
	})

	if gp.eventBus != nil && oldOxygen != newOxygen {
		events.Publish(gp.eventBus, events.OxygenChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOxygen,
			NewValue: newOxygen,
		})
	}

	return nil
}

func (gp *GlobalParameters) SetOceans(ctx context.Context, newOceans int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldOceans int
	gp.update(func(s *datastore.GameState) {
		oldOceans = s.Oceans
		s.Oceans = newOceans
	})

	if gp.eventBus != nil && oldOceans != newOceans {
		events.Publish(gp.eventBus, events.OceansChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldOceans,
			NewValue: newOceans,
		})
	}

	return nil
}

// IncreaseVenus raises the venus level by the specified number of steps.
// Each step is 2%. Returns the actual number of steps raised.
func (gp *GlobalParameters) IncreaseVenus(ctx context.Context, steps int, playerID string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var oldVenus, newVenus int
	gp.update(func(s *datastore.GameState) {
		oldVenus = s.Venus
		newVenus = min(s.Venus+steps*2, MaxVenus)
		s.Venus = newVenus
	})
	actualSteps := (newVenus - oldVenus) / 2

	if gp.eventBus != nil && oldVenus != newVenus {
		events.Publish(gp.eventBus, events.VenusChangedEvent{
			GameID:    gp.gameID,
			OldValue:  oldVenus,
			NewValue:  newVenus,
			ChangedBy: playerID,
		})
	}

	return actualSteps, nil
}

func (gp *GlobalParameters) SetVenus(ctx context.Context, newVenus int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldVenus int
	gp.update(func(s *datastore.GameState) {
		oldVenus = s.Venus
		s.Venus = newVenus
	})

	if gp.eventBus != nil && oldVenus != newVenus {
		events.Publish(gp.eventBus, events.VenusChangedEvent{
			GameID:   gp.gameID,
			OldValue: oldVenus,
			NewValue: newVenus,
		})
	}

	return nil
}
