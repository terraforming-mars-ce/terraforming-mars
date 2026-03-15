package core_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

func TestCardWarnings_TemperatureOutputWarnsWhenMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-temp-card",
		Name: "Temperature Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTemperature, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for temperature, got: %+v", state.Warnings)
	}
}

func TestCardWarnings_OxygenOutputWarnsWhenMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetOxygen(ctx, global_parameters.MaxOxygen); err != nil {
		t.Fatalf("Failed to set oxygen: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-oxygen-card",
		Name: "Oxygen Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOxygen, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for oxygen, got: %+v", state.Warnings)
	}
}

func TestCardWarnings_VenusOutputWarnsWhenMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetVenus(ctx, global_parameters.MaxVenus); err != nil {
		t.Fatalf("Failed to set venus: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-venus-card",
		Name: "Venus Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for venus, got: %+v", state.Warnings)
	}
}

func TestCardWarnings_NoWarningBelowMax(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetTemperature(ctx, 0); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-temp-card",
		Name: "Temperature Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTemperature, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	if len(state.Warnings) != 0 {
		t.Errorf("Expected no warnings below max, got: %+v", state.Warnings)
	}
}

func TestCardWarnings_WarningsDoNotPreventAvailability(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-temp-card",
		Name: "Temperature Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTemperature, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	if len(state.Warnings) == 0 {
		t.Fatal("Expected warnings to be present")
	}
	if !state.Available() {
		t.Errorf("Card should still be available despite warnings, errors: %+v", state.Errors)
	}
}

func TestCardWarnings_GreeneryTileWarnsWhenOxygenMaxed(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetOxygen(ctx, global_parameters.MaxOxygen); err != nil {
		t.Fatalf("Failed to set oxygen: %v", err)
	}

	card := &gamecards.Card{
		ID:   "test-greenery-card",
		Name: "Greenery Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 5,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceGreeneryTile, Amount: 1},
				},
			},
		},
	}

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeNoTRGain {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected no-tr-gain warning for greenery when oxygen maxed, got: %+v", state.Warnings)
	}
}

func TestCardActionWarnings_TemperatureOutputWarnsWhenMaxed(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTemperature, Amount: 1},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for action temperature output, got: %+v", state.Warnings)
	}
}

func TestCardActionWarnings_NoWarningBelowMax(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetTemperature(ctx, 0); err != nil {
		t.Fatalf("Failed to set temperature: %v", err)
	}

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceTemperature, Amount: 1},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	if len(state.Warnings) != 0 {
		t.Errorf("Expected no warnings below max, got: %+v", state.Warnings)
	}
}

func TestCardActionWarnings_ChoiceOutputWarnsWhenMaxed(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)
	ctx := context.Background()

	if err := g.GlobalParameters().SetVenus(ctx, global_parameters.MaxVenus); err != nil {
		t.Fatalf("Failed to set venus: %v", err)
	}

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1},
				},
			},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	found := false
	for _, w := range state.Warnings {
		if w.Code == player.WarningCodeGlobalParamMaxed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected global-param-maxed warning for choice venus output, got: %+v", state.Warnings)
	}
}
