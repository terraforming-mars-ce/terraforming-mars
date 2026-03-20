package standardproject

import "terraforming-mars-backend/internal/game/shared"

// StandardProjectDefinition is the static template loaded from JSON
type StandardProjectDefinition struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Pack        string                `json:"pack"`
	Behaviors   []shared.CardBehavior `json:"behaviors"`
	Style       Style                 `json:"style"`
}

// CreditCost extracts the credit input amount from behaviors (the cost to execute this project)
func (d *StandardProjectDefinition) CreditCost() int {
	for _, b := range d.Behaviors {
		for _, input := range b.Inputs {
			if input.ResourceType == shared.ResourceCredit {
				return input.Amount
			}
		}
	}
	return 0
}

// Style provides visual hints for the frontend
type Style struct {
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
