package shared

// PaymentSubstitute represents a resource type substitution for payments
type PaymentSubstitute struct {
	ResourceType   ResourceType
	ConversionRate int
}

// StoragePaymentSubstitute represents card storage resources that can be used as M€ payment.
// e.g., Dirigibles: floaters worth 3 M€ each when playing Venus-tagged cards.
type StoragePaymentSubstitute struct {
	CardID         string       `json:"cardId"`
	ResourceType   ResourceType `json:"resourceType"`
	ConversionRate int          `json:"conversionRate"`
	Selectors      []Selector   `json:"selectors,omitempty"` // What cards this applies to (e.g., Venus-tagged)
}

// RequirementModifier represents a modification to requirements
type RequirementModifier struct {
	Amount                int
	AffectedResources     []ResourceType
	CardTarget            *string
	StandardProjectTarget *StandardProject
}
