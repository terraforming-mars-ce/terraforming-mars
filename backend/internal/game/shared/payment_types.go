package shared

// PaymentSource identifies the physical resource pool used to pay a cost.
// CardID is bound at runtime and must be empty in card definitions.
type PaymentSource struct {
	Target   string       `json:"target"`
	Resource ResourceType `json:"resource"`
	CardID   string       `json:"cardId,omitempty"`
}

// PaymentSubstitute grants a direct exchange for eligible payments.
type PaymentSubstitute struct {
	Source          PaymentSource `json:"source"`
	TargetResource  ResourceType  `json:"targetResource"`
	ConversionRate  int           `json:"conversionRate"`
	GrantedByCardID string        `json:"grantedByCardId,omitempty"`
	Selectors       []Selector    `json:"selectors,omitempty"`
}

// PaymentAllocation spends units of one source toward one cost resource.
type PaymentAllocation struct {
	Source         PaymentSource `json:"source"`
	TargetResource ResourceType  `json:"targetResource"`
	Amount         int           `json:"amount"`
}

// Payment contains quantities selected by the payer, never client-provided rates.
type Payment struct {
	Allocations []PaymentAllocation `json:"allocations"`
}

// PaymentOption is a server-resolved exchange available for a cost.
type PaymentOption struct {
	Source         PaymentSource `json:"source"`
	TargetResource ResourceType  `json:"targetResource"`
	ConversionRate int           `json:"conversionRate"`
	Available      int           `json:"available"`
}

// PaymentQuote describes effective costs and the legal sources for those costs.
type PaymentQuote struct {
	Costs   map[ResourceType]int `json:"costs"`
	Options []PaymentOption      `json:"options"`
}

// NativePayment selects a resource from the player's own pool.
func NativePayment(resource ResourceType, amount int) Payment {
	if amount == 0 {
		return Payment{Allocations: []PaymentAllocation{}}
	}
	return Payment{Allocations: []PaymentAllocation{{Source: PaymentSource{Target: "self-player", Resource: resource}, TargetResource: resource, Amount: amount}}}
}

// RequirementModifier represents a modification to requirements
type RequirementModifier struct {
	Amount                int
	AffectedResources     []ResourceType
	CardTarget            *string
	StandardProjectTarget *StandardProject
}

// IsBasicPaymentResource reports whether a resource has a spendable player pool.
func IsBasicPaymentResource(rt ResourceType) bool {
	switch rt {
	case ResourceCredit, ResourceSteel, ResourceTitanium, ResourcePlant, ResourceEnergy, ResourceHeat:
		return true
	}
	return false
}
