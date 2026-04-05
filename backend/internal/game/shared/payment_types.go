package shared

// PaymentSubstitute represents a resource type substitution for payments
type PaymentSubstitute struct {
	ResourceType   ResourceType
	ConversionRate int
}

// StoragePaymentSubstitute represents card storage resources that can be used as payment.
// TargetResource indicates what the storage converts into (e.g., credit for Dirigibles, heat for Storm Craft).
type StoragePaymentSubstitute struct {
	CardID         string       `json:"cardId"`
	ResourceType   ResourceType `json:"resourceType"`
	ConversionRate int          `json:"conversionRate"`
	TargetResource ResourceType `json:"targetResource"`
	Selectors      []Selector   `json:"selectors,omitempty"`
}

// RequirementModifier represents a modification to requirements
type RequirementModifier struct {
	Amount                int
	AffectedResources     []ResourceType
	CardTarget            *string
	StandardProjectTarget *StandardProject
}
