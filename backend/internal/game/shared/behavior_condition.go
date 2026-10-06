package shared

// BehaviorCondition is the interface for all resource condition types (inputs and outputs).
// Implemented by typed category structs (BasicResourceCondition, ProductionCondition, etc.).
type BehaviorCondition interface {
	GetResourceType() ResourceType
	GetAmount() int
	GetTarget() string
	SetAmount(int)
	deepCopyCondition() BehaviorCondition
	isBehaviorCondition()
}

// ConditionBase holds the common fields shared by all condition categories.
// Category-specific structs embed this.
type ConditionBase struct {
	ResourceType ResourceType `json:"type"`
	Amount       int          `json:"amount"`
	Target       string       `json:"target"`
}

func (b *ConditionBase) GetResourceType() ResourceType { return b.ResourceType }
func (b *ConditionBase) GetAmount() int                { return b.Amount }
func (b *ConditionBase) GetTarget() string             { return b.Target }
func (b *ConditionBase) SetAmount(a int)               { b.Amount = a }

// CloneCondition creates a deep copy of a BehaviorCondition.
func CloneCondition(bc BehaviorCondition) BehaviorCondition {
	return bc.deepCopyCondition()
}

// GetPerCondition extracts the Per field from a typed condition, or nil.
func GetPerCondition(bc BehaviorCondition) *PerCondition {
	switch c := bc.(type) {
	case *CardOperationCondition:
		return c.Per
	case *BasicResourceCondition:
		return c.Per
	case *ProductionCondition:
		return c.Per
	case *GlobalParameterCondition:
		return c.Per
	case *CardStorageCondition:
		return c.Per
	case *MiscCondition:
		return c.Per
	default:
		return nil
	}
}

// CalculateScaledAmount applies a per-condition divisor and caps matched groups before multiplying.
// Conditions without a positive divisor retain their base amount.
func CalculateScaledAmount(output BehaviorCondition, count int) int {
	per := GetPerCondition(output)
	if per == nil || per.Amount <= 0 {
		return output.GetAmount()
	}
	groups := count / per.Amount
	if basic, ok := output.(*BasicResourceCondition); ok && basic.MaxTrigger != nil {
		groups = min(groups, *basic.MaxTrigger)
	}
	return output.GetAmount() * groups
}

// IsVariableAmount returns whether a condition has VariableAmount set.
func IsVariableAmount(bc BehaviorCondition) bool {
	switch c := bc.(type) {
	case *BasicResourceCondition:
		return c.VariableAmount
	case *ProductionCondition:
		return c.VariableAmount
	case *CardStorageCondition:
		return c.VariableAmount
	case *CardOperationCondition:
		return c.VariableAmount
	default:
		return false
	}
}

// IsOptional returns whether a condition has Optional set.
func IsOptional(bc BehaviorCondition) bool {
	switch c := bc.(type) {
	case *ColonyCondition:
		return c.Optional
	case *BasicResourceCondition:
		return c.Optional
	case *TilePlacementCondition:
		return c.Optional
	case *CardOperationCondition:
		return c.Optional
	case *CardStorageCondition:
		return c.Optional
	default:
		return false
	}
}

// GetPaymentAllowed extracts PaymentAllowed from a typed condition, or nil.
func GetPaymentAllowed(bc BehaviorCondition) []ResourceType {
	switch c := bc.(type) {
	case *BasicResourceCondition:
		return c.PaymentAllowed
	case *CardOperationCondition:
		return c.PaymentAllowed
	default:
		return nil
	}
}

// GetTemporary extracts the Temporary field from a typed condition, or empty string.
func GetTemporary(bc BehaviorCondition) string {
	switch c := bc.(type) {
	case *EffectCondition:
		return c.Temporary
	default:
		return ""
	}
}

// GetSelectors extracts the Selectors field from a typed condition, or nil.
func GetSelectors(bc BehaviorCondition) []Selector {
	switch c := bc.(type) {
	case *PaymentSubstituteCondition:
		return c.Selectors
	case *CopyCondition:
		return c.Selectors
	case *CardOperationCondition:
		return c.Selectors
	case *CardStorageCondition:
		return c.Selectors
	case *EffectCondition:
		return c.Selectors
	case *MiscCondition:
		return c.Selectors
	default:
		return nil
	}
}

// GetTileRestrictions extracts the TileRestrictions field from a typed condition, or nil.
func GetTileRestrictions(bc BehaviorCondition) *TileRestrictions {
	switch c := bc.(type) {
	case *TilePlacementCondition:
		return c.TileRestrictions
	default:
		return nil
	}
}
