package behavior_test

import (
	"context"
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// applyOutputs applies outputs directly via BehaviorApplier without card play.
func applyOutputs(t *testing.T, p *player.Player, g *game.Game, registry cards.CardRegistry, outputs ...shared.BehaviorCondition) {
	t.Helper()
	applier := gamecards.NewBehaviorApplier(p, g, "test", testutil.TestLogger()).
		WithCardRegistry(registry)
	err := applier.ApplyOutputs(context.Background(), outputs)
	testutil.AssertNoError(t, err, "ApplyOutputs failed")
}

// applyInputs applies inputs directly via BehaviorApplier without card play.
func applyInputs(t *testing.T, p *player.Player, g *game.Game, inputs ...shared.BehaviorCondition) {
	t.Helper()
	applier := gamecards.NewBehaviorApplier(p, g, "test", testutil.TestLogger())
	err := applier.ApplyInputs(context.Background(), inputs)
	testutil.AssertNoError(t, err, "ApplyInputs failed")
}

// applyOptions configures the BehaviorApplier for tests needing extra context.
type applyOptions struct {
	sourceCardID      string
	targetCardIDs     []string
	targetPlayerID    string
	stealSourceCardID string
	selectedAmount    int
	cardRegistry      cards.CardRegistry
}

// applyOutputsWithOptions applies outputs with additional applier configuration.
func applyOutputsWithOptions(t *testing.T, p *player.Player, g *game.Game, opts applyOptions, outputs ...shared.BehaviorCondition) {
	t.Helper()
	applier := gamecards.NewBehaviorApplier(p, g, "test", testutil.TestLogger())
	if opts.sourceCardID != "" {
		applier = applier.WithSourceCardID(opts.sourceCardID)
	}
	if len(opts.targetCardIDs) > 0 {
		applier = applier.WithTargetCardIDs(opts.targetCardIDs)
	}
	if opts.targetPlayerID != "" {
		applier = applier.WithTargetPlayerID(opts.targetPlayerID)
	}
	if opts.stealSourceCardID != "" {
		applier = applier.WithStealSourceCardID(opts.stealSourceCardID)
	}
	if opts.selectedAmount != 0 {
		applier = applier.WithSelectedAmount(opts.selectedAmount)
	}
	if opts.cardRegistry != nil {
		applier = applier.WithCardRegistry(opts.cardRegistry)
	}
	err := applier.ApplyOutputs(context.Background(), outputs)
	testutil.AssertNoError(t, err, "ApplyOutputs failed")
}

// assertResources checks multiple resource amounts on a player.
func assertResources(t *testing.T, p *player.Player, expected map[shared.ResourceType]int) {
	t.Helper()
	resources := p.Resources().Get()
	for rt, want := range expected {
		var got int
		switch rt {
		case shared.ResourceCredit:
			got = resources.Credits
		case shared.ResourceSteel:
			got = resources.Steel
		case shared.ResourceTitanium:
			got = resources.Titanium
		case shared.ResourcePlant:
			got = resources.Plants
		case shared.ResourceEnergy:
			got = resources.Energy
		case shared.ResourceHeat:
			got = resources.Heat
		default:
			t.Fatalf("assertResources: unsupported resource type %s", rt)
		}
		testutil.AssertEqual(t, want, got, fmt.Sprintf("expected %d %s, got %d", want, rt, got))
	}
}

// assertProduction checks multiple production amounts on a player.
func assertProduction(t *testing.T, p *player.Player, expected map[shared.ResourceType]int) {
	t.Helper()
	production := p.Resources().Production()
	for rt, want := range expected {
		got := production.GetAmount(rt)
		testutil.AssertEqual(t, want, got, fmt.Sprintf("expected %d %s production, got %d", want, rt, got))
	}
}

// makeTestCard creates a minimal card definition for play-card tests.
func makeTestCard(id, name string, cost int, behaviors ...shared.CardBehavior) gamecards.Card {
	return gamecards.Card{
		ID:        id,
		Name:      name,
		Type:      gamecards.CardTypeAutomated,
		Cost:      cost,
		Behaviors: behaviors,
	}
}

// autoBehavior creates an auto-trigger behavior with the given outputs.
func autoBehavior(outputs ...shared.BehaviorCondition) shared.CardBehavior {
	return shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
		Outputs:  outputs,
	}
}
