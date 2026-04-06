package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyEffectOutput(ctx context.Context, o *shared.EffectCondition, amount int, log *zap.Logger) error {
	switch o.ResourceType {
	case shared.ResourcePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply payment substitute: no player context")
		}
		resources := GetResourcesFromSelectors(o.Selectors)
		if len(resources) > 0 {
			resourceType := shared.ResourceType(resources[0])
			a.player.Resources().AddPaymentSubstitute(resourceType, amount)
			log.Debug("Added payment substitute",
				zap.String("resource_type", string(resourceType)), zap.Int("conversion_rate", amount))
		} else {
			log.Warn("payment-substitute output missing selectors with resources")
		}

	case shared.ResourceDiscount:
		log.Debug("Discount effect registered", zap.Int("amount", amount), zap.Any("selectors", o.Selectors))

	case shared.ResourceGlobalParameterLenience:
		log.Debug("Global parameter lenience effect registered", zap.Int("amount", amount), zap.String("temporary", o.Temporary))

	case shared.ResourceIgnoreGlobalRequirements:
		log.Debug("Ignore global requirements effect registered", zap.String("temporary", o.Temporary))

	case shared.ResourceValueModifier:
		if a.player == nil {
			return fmt.Errorf("cannot apply value modifier: no player context")
		}
		for _, resourceStr := range GetResourcesFromSelectors(o.Selectors) {
			resourceType := shared.ResourceType(resourceStr)
			a.player.Resources().AddValueModifier(resourceType, amount)
			log.Debug("Added resource value modifier",
				zap.String("resource_type", string(resourceType)), zap.Int("modifier_amount", amount))
		}

	case shared.ResourceStoragePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply storage payment substitute: no player context")
		}
		if a.sourceCardID == "" {
			log.Warn("storage-payment-substitute output missing source card ID")
			return nil
		}
		storageResourceType := shared.ResourceFloater
		if a.cardRegistry != nil {
			if sourceCard, err := a.cardRegistry.GetByID(a.sourceCardID); err == nil && sourceCard.ResourceStorage != nil {
				storageResourceType = sourceCard.ResourceStorage.Type
			}
		}
		targetResource := shared.ResourceCredit
		resources := GetResourcesFromSelectors(o.Selectors)
		if len(resources) > 0 {
			targetResource = shared.ResourceType(resources[0])
		}
		a.player.Resources().AddStoragePaymentSubstitute(shared.StoragePaymentSubstitute{
			CardID:         a.sourceCardID,
			ResourceType:   storageResourceType,
			ConversionRate: amount,
			TargetResource: targetResource,
			Selectors:      o.Selectors,
		})
		log.Debug("Added storage payment substitute",
			zap.String("card_id", a.sourceCardID), zap.String("resource_type", string(storageResourceType)),
			zap.String("target_resource", string(targetResource)), zap.Int("conversion_rate", amount))

	case shared.ResourceOceanAdjacencyBonus:
		log.Debug("Ocean adjacency bonus effect registered", zap.Int("amount", amount))

	case shared.ResourceDefense:
		log.Debug("Defense effect registered", zap.Int("amount", amount), zap.Any("selectors", o.Selectors))

	case shared.ResourceActionReuse:
		log.Debug("Skipping action-reuse output (handled at action layer)")

	case shared.ResourceEffect:
		log.Debug("Effect registered", zap.Int("amount", amount))

	case shared.ResourceTag:
		log.Debug("Tag effect registered", zap.Int("amount", amount))

	default:
		log.Warn("Unhandled effect type", zap.String("type", string(o.ResourceType)))
	}
	return nil
}
