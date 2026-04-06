package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyBasicResourceOutput(ctx context.Context, o *shared.BasicResourceCondition, amount int, log *zap.Logger) error {
	rt := o.ResourceType

	// Special case: credit steal with adjacency restriction (deferred for post-tile-placement)
	if rt == shared.ResourceCredit && o.Target == "steal-any-player" && o.TargetRestriction != nil && o.TargetRestriction.Adjacent == "self-card" {
		a.deferredSteal = o
		log.Debug("Deferred adjacent steal for post-tile-placement", zap.Int("amount", amount))
		return nil
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
	log.Debug("Added resource", zap.String("type", string(rt)), zap.Int("amount", amount))
	return nil
}

func (a *BehaviorApplier) applyProductionOutput(ctx context.Context, o *shared.ProductionCondition, amount int, log *zap.Logger) error {
	rt := o.ResourceType
	if o.Target == "any-player" {
		return a.applyAnyPlayerProduction(rt, amount, log)
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply %s: no player context", rt)
	}
	a.player.Resources().AddProduction(map[shared.ResourceType]int{rt: amount})
	log.Debug("Added production", zap.String("type", string(rt)), zap.Int("amount", amount))
	return nil
}

func (a *BehaviorApplier) applyGlobalParameterOutput(ctx context.Context, o *shared.GlobalParameterCondition, amount int, log *zap.Logger) error {
	switch o.ResourceType {
	case shared.ResourceTR:
		if a.player == nil {
			return fmt.Errorf("cannot apply terraform rating: no player context")
		}
		a.player.Resources().UpdateTerraformRating(amount)
		log.Debug("Added terraform rating", zap.Int("amount", amount))

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
		log.Debug("Increased oxygen", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

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
		log.Debug("Increased temperature", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

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
		log.Debug("Increased venus", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

	default:
		log.Warn("Unhandled global parameter type", zap.String("type", string(o.ResourceType)))
	}
	return nil
}
