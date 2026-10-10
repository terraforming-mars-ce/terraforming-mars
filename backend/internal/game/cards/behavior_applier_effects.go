package cards

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game/shared"
)

func (a *BehaviorApplier) applyEffectOutput(ctx context.Context, o *shared.EffectCondition, amount int, log *slog.Logger) error {
	switch o.ResourceType {

	case shared.ResourceDiscount:
		log.Debug("Discount effect registered", slog.Int("amount", amount), slog.Any("selectors", o.Selectors))

	case shared.ResourceGlobalParameterLenience:
		log.Debug("Global parameter lenience effect registered", slog.Int("amount", amount), slog.String("temporary", o.Temporary))

	case shared.ResourceIgnoreGlobalRequirements:
		log.Debug("Ignore global requirements effect registered", slog.String("temporary", o.Temporary))

	case shared.ResourceValueModifier:
		if a.player == nil {
			return fmt.Errorf("cannot apply value modifier: no player context")
		}
		for _, resourceStr := range GetResourcesFromSelectors(o.Selectors) {
			resourceType := shared.ResourceType(resourceStr)
			a.player.Resources().AddValueModifier(resourceType, amount)
			log.Debug("Added resource value modifier",
				slog.String("resource_type", string(resourceType)), slog.Int("modifier_amount", amount))
		}

	case shared.ResourceOceanAdjacencyBonus:
		log.Debug("Ocean adjacency bonus effect registered", slog.Int("amount", amount))

	case shared.ResourceDefense:
		log.Debug("Defense effect registered", slog.Int("amount", amount), slog.Any("selectors", o.Selectors))

	case shared.ResourceActionReuse:
		log.Debug("Skipping action-reuse output (handled at action layer)")

	case shared.ResourceEffect:
		log.Debug("Effect registered", slog.Int("amount", amount))

	case shared.ResourceTag:
		log.Debug("Tag effect registered", slog.Int("amount", amount))

	default:
		log.Warn("Unhandled effect type", slog.String("type", string(o.ResourceType)))
	}
	return nil
}
