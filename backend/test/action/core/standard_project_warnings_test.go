package core_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// TestStandardProject_AsteroidWarnsWhenTemperatureMaxed verifies a warning is returned
// when the Asteroid standard project is used but temperature is already at max.
func TestStandardProject_AsteroidWarnsWhenTemperatureMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	if err := g.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature); err != nil {
		t.Fatalf("Failed to set temperature to max: %v", err)
	}

	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectAsteroid,
		p,
		g,
		cardRegistry,
	)

	// Should still be available (warning, not error)
	if !state.Available() {
		t.Errorf("Asteroid should be available even at max temperature, got errors: %+v", state.Errors)
	}

	// Should have a warning
	if len(state.Warnings) == 0 {
		t.Fatal("Expected a warning for maxed temperature")
	}

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning, got warnings: %+v", state.Warnings)
	}
}

// TestStandardProject_AsteroidNoWarningBelowMax verifies no warning when temperature is not maxed.
func TestStandardProject_AsteroidNoWarningBelowMax(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	if err := g.GlobalParameters().SetTemperature(ctx, 0); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectAsteroid,
		p,
		g,
		cardRegistry,
	)

	if len(state.Warnings) != 0 {
		t.Errorf("Expected no warnings below max temperature, got: %+v", state.Warnings)
	}
}

// TestStandardProject_GreeneryWarnsWhenOxygenMaxed verifies a warning is returned
// when the Greenery standard project is used but oxygen is already at max.
func TestStandardProject_GreeneryWarnsWhenOxygenMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	if err := g.GlobalParameters().SetOxygen(ctx, global_parameters.MaxOxygen); err != nil {
		t.Fatalf("Failed to set oxygen to max: %v", err)
	}

	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectGreenery,
		p,
		g,
		cardRegistry,
	)

	// Should still be available (warning, not error) — assuming placements exist
	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for maxed oxygen, got warnings: %+v", state.Warnings)
	}
}
