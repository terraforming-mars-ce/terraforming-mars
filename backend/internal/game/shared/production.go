package shared

const (
	// MinCreditProduction is the minimum MC production allowed (-5 in TM rules)
	MinCreditProduction = -5
	// MinOtherProduction is the minimum production for non-MC resources
	MinOtherProduction = 0
)

// Production represents a player's production values
type Production struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// DeepCopy creates a deep copy of the Production struct
func (p Production) DeepCopy() Production {
	return Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}

// GetAmount returns the production amount for a specific resource type.
// Accepts both base resources (e.g., "titanium") and production types (e.g., "titanium-production").
func (p Production) GetAmount(resourceType ResourceType) int {
	switch resourceType {
	case ResourceCredit, ResourceCreditProduction:
		return p.Credits
	case ResourceSteel, ResourceSteelProduction:
		return p.Steel
	case ResourceTitanium, ResourceTitaniumProduction:
		return p.Titanium
	case ResourcePlant, ResourcePlantProduction:
		return p.Plants
	case ResourceEnergy, ResourceEnergyProduction:
		return p.Energy
	case ResourceHeat, ResourceHeatProduction:
		return p.Heat
	default:
		return 0
	}
}
