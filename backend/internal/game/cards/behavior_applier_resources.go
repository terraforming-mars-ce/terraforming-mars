package cards

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game/shared"
)

func (a *BehaviorApplier) applyBasicResourceOutput(ctx context.Context, o *shared.BasicResourceCondition, amount int, log *slog.Logger) error {
	rt := o.ResourceType

	if o.Target == "any-player" && amount < 0 && o.TargetRestriction != nil {
		output := shared.CloneCondition(o).(*shared.BasicResourceCondition)
		output.Amount = amount
		if output.TargetRestriction.Adjacent != "" {
			a.deferredRemoval = output
			return nil
		}
		return QueueResourceRemoval(a.game, a.player, output, nil, a.sourceCardID, a.source, a.cardRegistry)
	}

	if o.Target == "steal-any-player" {
		return a.stealAnyPlayerResource(rt, amount, log)
	}
	if o.Target == "any-player" {
		return a.applyAnyPlayerResource(rt, amount, log)
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply %s: no player context", rt)
	}
	a.player.Resources().Add(map[shared.ResourceType]int{rt: amount})
	log.Debug("Added resource", slog.String("type", string(rt)), slog.Int("amount", amount))
	return nil
}

func (a *BehaviorApplier) applyProductionOutput(ctx context.Context, o *shared.ProductionCondition, amount int, log *slog.Logger) error {
	rt := o.ResourceType
	if a.productionBox == "resolved" && a.player != nil {
		a.player.Resources().RecordProductionBox(a.sourceCardID, *shared.NewProductionCondition(rt, amount, o.Target))
	}
	if o.Target == "any-player" {
		return a.applyAnyPlayerProduction(rt, amount, log)
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply %s: no player context", rt)
	}
	a.player.Resources().AddProduction(map[shared.ResourceType]int{rt: amount})
	log.Debug("Added production", slog.String("type", string(rt)), slog.Int("amount", amount))
	return nil
}

func (a *BehaviorApplier) applyGlobalParameterOutput(ctx context.Context, o *shared.GlobalParameterCondition, amount int, log *slog.Logger) error {
	switch o.ResourceType {
	case shared.ResourceTR:
		if a.player == nil {
			return fmt.Errorf("cannot apply terraform rating: no player context")
		}
		a.player.Resources().UpdateTerraformRating(amount)
		log.Debug("Added terraform rating", slog.Int("amount", amount))

	case shared.ResourceOxygen:
		if a.game == nil {
			return fmt.Errorf("cannot apply oxygen: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseOxygen(ctx, amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased oxygen", slog.Int("steps", actualSteps), slog.Int("tr_gained", actualSteps))

	case shared.ResourceTemperature:
		if a.game == nil {
			return fmt.Errorf("cannot apply temperature: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseTemperature(ctx, amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase temperature: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased temperature", slog.Int("steps", actualSteps), slog.Int("tr_gained", actualSteps))

	case shared.ResourceVenus:
		if a.game == nil {
			return fmt.Errorf("cannot apply venus: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseVenus(ctx, amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase venus: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased venus", slog.Int("steps", actualSteps), slog.Int("tr_gained", actualSteps))

	default:
		log.Warn("Unhandled global parameter type", slog.String("type", string(o.ResourceType)))
	}
	return nil
}
