package shared

// Choice policy constants
const (
	ChoicePolicyLowest         = "lowest"            // Only choices affecting the lowest production are valid
	ChoicePolicyAutoPlantTags3 = "auto-plant-tags-3" // Auto-select: choice 1 if 3+ plant tags, else choice 0
)

// CardBehavior represents card behaviors (immediate and repeatable)
type CardBehavior struct {
	Description                   string                         `json:"description,omitempty" ts:"string | undefined"`
	Triggers                      []Trigger                      `json:"triggers,omitempty"`
	Inputs                        []ResourceCondition            `json:"inputs,omitempty"`
	Outputs                       []ResourceCondition            `json:"outputs,omitempty"`
	Choices                       []Choice                       `json:"choices,omitempty"`
	ChoicePolicy                  string                         `json:"choicePolicy,omitempty"`
	GenerationalEventRequirements []GenerationalEventRequirement `json:"generationalEventRequirements,omitempty" ts:"GenerationalEventRequirement[] | undefined"`
	Group                         string                         `json:"group,omitempty" ts:"string | undefined"`
}

// DeepCopy creates a deep copy of the CardBehavior
func (cb CardBehavior) DeepCopy() CardBehavior {
	var result CardBehavior

	result.Description = cb.Description
	result.ChoicePolicy = cb.ChoicePolicy
	result.Group = cb.Group

	if cb.Triggers != nil {
		result.Triggers = make([]Trigger, len(cb.Triggers))
		for i, trigger := range cb.Triggers {
			result.Triggers[i] = trigger
		}
	}

	if cb.Inputs != nil {
		result.Inputs = make([]ResourceCondition, len(cb.Inputs))
		for i, input := range cb.Inputs {
			result.Inputs[i] = deepCopyResourceCondition(input)
		}
	}

	if cb.Outputs != nil {
		result.Outputs = make([]ResourceCondition, len(cb.Outputs))
		for i, output := range cb.Outputs {
			result.Outputs[i] = deepCopyResourceCondition(output)
		}
	}

	if cb.Choices != nil {
		result.Choices = make([]Choice, len(cb.Choices))
		for i, choice := range cb.Choices {
			choiceCopy := Choice{}

			if choice.Inputs != nil {
				choiceCopy.Inputs = make([]ResourceCondition, len(choice.Inputs))
				for j, input := range choice.Inputs {
					choiceCopy.Inputs[j] = deepCopyResourceCondition(input)
				}
			}

			if choice.Outputs != nil {
				choiceCopy.Outputs = make([]ResourceCondition, len(choice.Outputs))
				for j, output := range choice.Outputs {
					choiceCopy.Outputs[j] = deepCopyResourceCondition(output)
				}
			}

			if choice.Requirements != nil {
				items := make([]ChoiceRequirement, len(choice.Requirements.Items))
				copy(items, choice.Requirements.Items)
				choiceCopy.Requirements = &ChoiceRequirements{Items: items}
			}

			result.Choices[i] = choiceCopy
		}
	}

	if cb.GenerationalEventRequirements != nil {
		result.GenerationalEventRequirements = make([]GenerationalEventRequirement, len(cb.GenerationalEventRequirements))
		for i, req := range cb.GenerationalEventRequirements {
			result.GenerationalEventRequirements[i] = deepCopyGenerationalEventRequirement(req)
		}
	}

	return result
}

// ExtractInputsOutputs extracts the combined inputs and outputs for a behavior,
// optionally incorporating a selected choice. Returns base + choice inputs/outputs.
// If choiceIndex is nil or out of range, only base inputs/outputs are returned.
func (cb CardBehavior) ExtractInputsOutputs(choiceIndex *int) (inputs []ResourceCondition, outputs []ResourceCondition) {
	if len(cb.Inputs) > 0 {
		inputs = make([]ResourceCondition, len(cb.Inputs))
		copy(inputs, cb.Inputs)
	}
	if len(cb.Outputs) > 0 {
		outputs = make([]ResourceCondition, len(cb.Outputs))
		copy(outputs, cb.Outputs)
	}

	if choiceIndex != nil && *choiceIndex >= 0 && *choiceIndex < len(cb.Choices) {
		selectedChoice := cb.Choices[*choiceIndex]

		// Append choice inputs to base inputs
		if len(selectedChoice.Inputs) > 0 {
			inputs = append(inputs, selectedChoice.Inputs...)
		}

		// Append choice outputs to base outputs
		if len(selectedChoice.Outputs) > 0 {
			outputs = append(outputs, selectedChoice.Outputs...)
		}
	}

	return inputs, outputs
}

// productionResourceTypes lists all production resource types for policy evaluation.
var productionResourceTypes = []ResourceType{
	ResourceCreditProduction,
	ResourceSteelProduction,
	ResourceTitaniumProduction,
	ResourcePlantProduction,
	ResourceEnergyProduction,
	ResourceHeatProduction,
}

// FilterChoiceIndicesByPolicy returns the indices of choices that are valid under the given policy.
// For empty policy or "any", all indices are returned.
// For "lowest", only choices whose production output is at the player's minimum production level.
func FilterChoiceIndicesByPolicy(choices []Choice, policy string, production Production) []int {
	if policy == "" {
		indices := make([]int, len(choices))
		for i := range choices {
			indices[i] = i
		}
		return indices
	}

	switch policy {
	case ChoicePolicyLowest:
		return filterLowestProductionChoices(choices, production)
	case ChoicePolicyAutoPlantTags3:
		// Handled externally via AutoSelectChoiceIndex — return all as valid fallback
		return allChoiceIndices(choices)
	default:
		indices := make([]int, len(choices))
		for i := range choices {
			indices[i] = i
		}
		return indices
	}
}

// IsChoiceValidForPolicy checks whether a specific choice index is valid under the given policy.
func IsChoiceValidForPolicy(choiceIndex int, choices []Choice, policy string, production Production) bool {
	if policy == "" || choiceIndex < 0 || choiceIndex >= len(choices) {
		return choiceIndex >= 0 && choiceIndex < len(choices)
	}

	validIndices := FilterChoiceIndicesByPolicy(choices, policy, production)
	for _, idx := range validIndices {
		if idx == choiceIndex {
			return true
		}
	}
	return false
}

func allChoiceIndices(choices []Choice) []int {
	indices := make([]int, len(choices))
	for i := range choices {
		indices[i] = i
	}
	return indices
}

// AutoSelectChoiceIndex returns an auto-selected choice index for policies that
// deterministically pick a choice based on game state. Returns -1 if not applicable.
func AutoSelectChoiceIndex(policy string, plantTagCount int) int {
	switch policy {
	case ChoicePolicyAutoPlantTags3:
		if plantTagCount >= 3 {
			return 1 // 4 plant production
		}
		return 0 // 1 plant production
	default:
		return -1
	}
}

func filterLowestProductionChoices(choices []Choice, production Production) []int {
	// Find the minimum production value across all 6 types
	minValue := production.Credits
	for _, rt := range productionResourceTypes {
		val := production.GetAmount(rt)
		if val < minValue {
			minValue = val
		}
	}

	// Return indices of choices whose production output type is at the minimum level
	var valid []int
	for i, choice := range choices {
		for _, output := range choice.Outputs {
			if isProductionResourceType(output.ResourceType) && production.GetAmount(output.ResourceType) == minValue {
				valid = append(valid, i)
				break
			}
		}
	}
	return valid
}

func isProductionResourceType(rt ResourceType) bool {
	switch rt {
	case ResourceCreditProduction, ResourceSteelProduction, ResourceTitaniumProduction,
		ResourcePlantProduction, ResourceEnergyProduction, ResourceHeatProduction:
		return true
	default:
		return false
	}
}

func deepCopyResourceCondition(rc ResourceCondition) ResourceCondition {
	result := rc

	if rc.Selectors != nil {
		result.Selectors = make([]Selector, len(rc.Selectors))
		copy(result.Selectors, rc.Selectors)
	}

	if rc.Per != nil {
		perCopy := *rc.Per
		result.Per = &perCopy
	}

	return result
}

func deepCopyGenerationalEventRequirement(req GenerationalEventRequirement) GenerationalEventRequirement {
	result := GenerationalEventRequirement{
		Event: req.Event,
	}

	if req.Count != nil {
		countCopy := MinMax{}
		if req.Count.Min != nil {
			minCopy := *req.Count.Min
			countCopy.Min = &minCopy
		}
		if req.Count.Max != nil {
			maxCopy := *req.Count.Max
			countCopy.Max = &maxCopy
		}
		result.Count = &countCopy
	}

	if req.Target != nil {
		targetCopy := *req.Target
		result.Target = &targetCopy
	}

	return result
}
