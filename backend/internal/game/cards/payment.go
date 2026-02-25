package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/shared"
)

// CardPayment represents how a player is paying for a card
type CardPayment struct {
	Credits            int
	Steel              int
	Titanium           int
	Substitutes        map[shared.ResourceType]int
	StorageSubstitutes map[string]int // cardID -> amount of storage resources to use as payment
}

// TotalValue calculates the total MC value of this payment.
// playerSubstitutes MUST include steel and titanium with their effective values
// (base value + any modifiers from cards like Phobolog or Advanced Alloys).
// All payment values (steel, titanium, other substitutes) are looked up from playerSubstitutes.
func (p CardPayment) TotalValue(playerSubstitutes []shared.PaymentSubstitute, storageSubstitutes []shared.StoragePaymentSubstitute) int {
	total := p.Credits

	substituteValues := make(map[shared.ResourceType]int)
	for _, sub := range playerSubstitutes {
		substituteValues[sub.ResourceType] = sub.ConversionRate
	}

	if steelValue, ok := substituteValues[shared.ResourceSteel]; ok {
		total += p.Steel * steelValue
	}

	if titaniumValue, ok := substituteValues[shared.ResourceTitanium]; ok {
		total += p.Titanium * titaniumValue
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			if conversionRate, ok := substituteValues[resourceType]; ok {
				total += amount * conversionRate
			}
		}
	}

	// Add storage substitute values (e.g., Dirigibles floaters at 3 M€ each)
	if p.StorageSubstitutes != nil {
		storageSubValues := make(map[string]int)
		for _, sub := range storageSubstitutes {
			storageSubValues[sub.CardID] = sub.ConversionRate
		}
		for cardID, amount := range p.StorageSubstitutes {
			if rate, ok := storageSubValues[cardID]; ok {
				total += amount * rate
			}
		}
	}

	return total
}

// Validate checks if the payment is valid
func (p CardPayment) Validate() error {
	if p.Credits < 0 {
		return fmt.Errorf("payment credits cannot be negative: %d", p.Credits)
	}
	if p.Steel < 0 {
		return fmt.Errorf("payment steel cannot be negative: %d", p.Steel)
	}
	if p.Titanium < 0 {
		return fmt.Errorf("payment titanium cannot be negative: %d", p.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			if amount < 0 {
				return fmt.Errorf("payment substitute %s cannot be negative: %d", resourceType, amount)
			}
		}
	}

	if p.StorageSubstitutes != nil {
		for cardID, amount := range p.StorageSubstitutes {
			if amount < 0 {
				return fmt.Errorf("storage payment substitute from card %s cannot be negative: %d", cardID, amount)
			}
		}
	}

	return nil
}

// CanAfford checks if a player has sufficient resources for this payment.
// storageGetter provides access to card storage amounts (cardID -> stored amount).
func (p CardPayment) CanAfford(playerResources shared.Resources, storageGetter func(cardID string) int) error {
	if playerResources.Credits < p.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", p.Credits, playerResources.Credits)
	}
	if playerResources.Steel < p.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", p.Steel, playerResources.Steel)
	}
	if playerResources.Titanium < p.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", p.Titanium, playerResources.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			var available int
			switch resourceType {
			case shared.ResourceHeat:
				available = playerResources.Heat
			case shared.ResourceEnergy:
				available = playerResources.Energy
			case shared.ResourcePlant:
				available = playerResources.Plants
			default:
				return fmt.Errorf("unsupported payment substitute resource type: %s", resourceType)
			}

			if available < amount {
				return fmt.Errorf("insufficient %s: need %d, have %d", resourceType, amount, available)
			}
		}
	}

	if p.StorageSubstitutes != nil && storageGetter != nil {
		for cardID, amount := range p.StorageSubstitutes {
			available := storageGetter(cardID)
			if available < amount {
				return fmt.Errorf("insufficient storage on card %s: need %d, have %d", cardID, amount, available)
			}
		}
	}

	return nil
}

// CoversCardCost checks if this payment covers the card cost
func (p CardPayment) CoversCardCost(cardCost int, allowSteel, allowTitanium bool, playerSubstitutes []shared.PaymentSubstitute, storageSubstitutes []shared.StoragePaymentSubstitute) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if p.Steel > 0 && !allowSteel {
		return fmt.Errorf("card does not have building tag, cannot use steel")
	}
	if p.Titanium > 0 && !allowTitanium {
		return fmt.Errorf("card does not have space tag, cannot use titanium")
	}

	if p.Substitutes != nil {
		for resourceType := range p.Substitutes {
			found := false
			for _, sub := range playerSubstitutes {
				if sub.ResourceType == resourceType {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("player cannot use %s as payment substitute", resourceType)
			}
		}
	}

	// Validate storage substitutes are from valid cards
	if p.StorageSubstitutes != nil {
		for cardID := range p.StorageSubstitutes {
			found := false
			for _, sub := range storageSubstitutes {
				if sub.CardID == cardID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("player cannot use storage from card %s as payment", cardID)
			}
		}
	}

	totalValue := p.TotalValue(playerSubstitutes, storageSubstitutes)
	if totalValue < cardCost {
		return fmt.Errorf("payment insufficient: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	return nil
}
