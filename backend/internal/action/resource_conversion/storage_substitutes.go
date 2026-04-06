package resource_conversion

import (
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ValidateAndDeductStorageSubstitutes validates that the requested storage substitutes are valid
// for the given target resource, deducts them from card storage, and returns the total resource
// value contributed by the substitutes.
func ValidateAndDeductStorageSubstitutes(
	p *player.Player,
	storageSubstitutes map[string]int,
	targetResource shared.ResourceType,
	log *zap.Logger,
) (int, error) {
	if len(storageSubstitutes) == 0 {
		return 0, nil
	}

	substituteLookup := make(map[string]shared.StoragePaymentSubstitute)
	for _, sub := range p.Resources().StoragePaymentSubstitutes() {
		if sub.TargetResource == targetResource {
			substituteLookup[sub.CardID] = sub
		}
	}

	totalValue := 0
	for cardID, count := range storageSubstitutes {
		if count <= 0 {
			continue
		}

		sub, exists := substituteLookup[cardID]
		if !exists {
			return 0, fmt.Errorf("no storage payment substitute targeting %s found for card %s", targetResource, cardID)
		}

		available := p.Resources().GetCardStorage(cardID)
		if available < count {
			return 0, fmt.Errorf("insufficient storage on card %s: need %d, have %d", cardID, count, available)
		}

		totalValue += count * sub.ConversionRate
		p.Resources().AddToStorage(cardID, -count)
		log.Debug("Deducted storage substitute",
			zap.String("card_id", cardID), zap.Int("count", count),
			zap.Int("value", count*sub.ConversionRate), zap.String("target", string(targetResource)))
	}

	return totalValue, nil
}

// CalculateStorageSubstituteValue returns the total available resource value from storage
// substitutes targeting the given resource type, without deducting anything.
func CalculateStorageSubstituteValue(
	p *player.Player,
	targetResource shared.ResourceType,
) int {
	totalValue := 0
	for _, sub := range p.Resources().StoragePaymentSubstitutes() {
		if sub.TargetResource == targetResource {
			stored := p.Resources().GetCardStorage(sub.CardID)
			totalValue += stored * sub.ConversionRate
		}
	}
	return totalValue
}
