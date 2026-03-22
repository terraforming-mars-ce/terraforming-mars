package cards

import (
	"slices"

	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CardLookup is a minimal interface for looking up cards by ID
// This avoids import cycles with internal/cards package
type CardLookup interface {
	GetByID(id string) (*Card, error)
}

// RequirementModifierCalculator computes requirement modifiers from player effects and hand
type RequirementModifierCalculator struct {
	cardLookup CardLookup
}

// NewRequirementModifierCalculator creates a new calculator
func NewRequirementModifierCalculator(cardLookup CardLookup) *RequirementModifierCalculator {
	return &RequirementModifierCalculator{
		cardLookup: cardLookup,
	}
}

// Calculate computes all requirement modifiers for a player based on their effects and hand
func (c *RequirementModifierCalculator) Calculate(p *player.Player) []shared.RequirementModifier {
	if p == nil {
		return []shared.RequirementModifier{}
	}

	modifiers := []shared.RequirementModifier{}

	effects := p.Effects().List()
	handCardIDs := p.Hand().Cards()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			// Check selectors for discount targeting
			if len(output.Selectors) > 0 {
				hasCardSelectors := HasCardSelectors(output.Selectors)
				hasStandardProjectSelectors := HasStandardProjectSelectors(output.Selectors)

				// Case 1: Standard project discount
				if hasStandardProjectSelectors {
					for _, selector := range output.Selectors {
						for _, project := range selector.StandardProjects {
							projectCopy := project
							affectedResources := c.convertAffectedResources(GetResourcesFromSelectors(output.Selectors))
							modifier := shared.RequirementModifier{
								Amount:                output.Amount,
								AffectedResources:     affectedResources,
								StandardProjectTarget: &projectCopy,
							}
							modifiers = append(modifiers, modifier)
						}
					}
				}

				// Case 2: Card discount (tag or type based)
				if hasCardSelectors {
					for _, cardID := range handCardIDs {
						card, err := c.cardLookup.GetByID(cardID)
						if err != nil {
							continue
						}

						if MatchesAnySelector(card, output.Selectors) {
							cardIDCopy := cardID
							modifier := shared.RequirementModifier{
								Amount:            output.Amount,
								AffectedResources: []shared.ResourceType{shared.ResourceCredit},
								CardTarget:        &cardIDCopy,
							}
							modifiers = append(modifiers, modifier)
						}
					}
				}
				continue
			}

			// Global discount (no selectors - applies to all cards in hand)
			for _, cardID := range handCardIDs {
				cardIDCopy := cardID
				modifier := shared.RequirementModifier{
					Amount:            output.Amount,
					AffectedResources: []shared.ResourceType{shared.ResourceCredit},
					CardTarget:        &cardIDCopy,
				}
				modifiers = append(modifiers, modifier)
			}
		}
	}

	return c.mergeModifiers(modifiers)
}

// convertAffectedResources converts string slice to ResourceType slice
func (c *RequirementModifierCalculator) convertAffectedResources(resources []string) []shared.ResourceType {
	if len(resources) == 0 {
		return []shared.ResourceType{shared.ResourceCredit} // Default to credits discount
	}
	result := make([]shared.ResourceType, len(resources))
	for i, r := range resources {
		result[i] = shared.ResourceType(r)
	}
	return result
}

// mergeModifiers combines modifiers targeting the same card/project by summing amounts
func (c *RequirementModifierCalculator) mergeModifiers(modifiers []shared.RequirementModifier) []shared.RequirementModifier {
	cardModifiers := make(map[string]*shared.RequirementModifier)
	projectModifiers := make(map[shared.StandardProject]*shared.RequirementModifier)

	for _, mod := range modifiers {
		if mod.CardTarget != nil {
			key := *mod.CardTarget
			if existing, ok := cardModifiers[key]; ok {
				existing.Amount += mod.Amount
			} else {
				modCopy := mod
				cardModifiers[key] = &modCopy
			}
		} else if mod.StandardProjectTarget != nil {
			key := *mod.StandardProjectTarget
			if existing, ok := projectModifiers[key]; ok {
				existing.Amount += mod.Amount
			} else {
				modCopy := mod
				projectModifiers[key] = &modCopy
			}
		}
	}

	result := make([]shared.RequirementModifier, 0, len(cardModifiers)+len(projectModifiers))
	for _, mod := range cardModifiers {
		result = append(result, *mod)
	}
	for _, mod := range projectModifiers {
		result = append(result, *mod)
	}
	return result
}

// CalculateCardDiscounts computes the total credit discount for a specific card.
// This is used during EntityState calculation instead of pre-computing all modifiers.
// Returns the total discount amount in credits that applies to this card.
func (c *RequirementModifierCalculator) CalculateCardDiscounts(p *player.Player, card *Card) int {
	if p == nil || card == nil {
		return 0
	}

	totalDiscount := 0
	effects := p.Effects().List()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			// Check selectors first (new system with AND logic within selector, OR between selectors)
			if len(output.Selectors) > 0 {
				hasCardSelectors := HasCardSelectors(output.Selectors)
				hasOnlyStandardProjectSelectors := HasStandardProjectSelectors(output.Selectors) && !hasCardSelectors

				if hasOnlyStandardProjectSelectors {
					continue
				}

				if hasCardSelectors {
					if MatchesAnySelector(card, output.Selectors) {
						totalDiscount += output.Amount
					}
					continue
				}

				totalDiscount += output.Amount
				continue
			}

			// Global discount (no selectors - applies to all cards)
			totalDiscount += output.Amount
		}
	}

	return totalDiscount
}

// CalculateGlobalParameterLenience computes the total lenience for a specific global parameter requirement.
// Lenience widens the min/max window: min is lowered, max is raised.
// The paramType should be one of: "temperature", "oxygen", "ocean", "venus".
func (c *RequirementModifierCalculator) CalculateGlobalParameterLenience(p *player.Player, paramType string) int {
	if p == nil {
		return 0
	}

	totalLenience := 0
	for _, effect := range p.Effects().List() {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceGlobalParameterLenience {
				continue
			}
			if len(output.Selectors) > 0 {
				if matchesGlobalParameterSelector(output.Selectors, paramType) {
					totalLenience += output.Amount
				}
			} else {
				totalLenience += output.Amount
			}
		}
	}

	return totalLenience
}

// HasIgnoreGlobalRequirements returns true if the player has an active effect
// that ignores all global parameter requirements (e.g. from Ecology Experts).
func (c *RequirementModifierCalculator) HasIgnoreGlobalRequirements(p *player.Player) bool {
	if p == nil {
		return false
	}
	for _, effect := range p.Effects().List() {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType == shared.ResourceIgnoreGlobalRequirements {
				return true
			}
		}
	}
	return false
}

func matchesGlobalParameterSelector(selectors []shared.Selector, paramType string) bool {
	for _, sel := range selectors {
		if slices.Contains(sel.GlobalParameters, paramType) {
			return true
		}
	}
	return false
}

// CalculateStandardProjectDiscounts computes discounts for a specific standard project.
// Returns a map of resource type to discount amount.
// For example, Ecoline's discount on PlantGreenery returns {"plants": 1}.
func (c *RequirementModifierCalculator) CalculateStandardProjectDiscounts(
	p *player.Player,
	projectType shared.StandardProject,
) map[shared.ResourceType]int {
	discounts := make(map[shared.ResourceType]int)

	if p == nil {
		return discounts
	}

	effects := p.Effects().List()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			// Check selectors for standard project discount
			if len(output.Selectors) > 0 {
				if MatchesAnyStandardProjectSelector(projectType, output.Selectors) {
					affectedResources := c.convertAffectedResources(GetResourcesFromSelectors(output.Selectors))
					for _, resourceType := range affectedResources {
						discounts[resourceType] += output.Amount
					}
				}
			}
		}
	}

	return discounts
}
