package cards

import (
	"slices"

	"terraforming-mars-backend/internal/game/shared"
)

// MatchesSelector checks if a card matches a single selector.
// Tags: ALL must be present (AND logic)
// CardTypes: ANY can match (OR logic - card has one type)
// Resources: card must have the resource (in storage, behavior outputs, or behavior inputs)
// RequiredOriginalCost: card cost must satisfy min/max constraints
// A selector must have at least one card-relevant criterion to match.
func MatchesSelector(card *Card, selector shared.Selector) bool {
	if len(selector.Tags) == 0 && len(selector.CardTypes) == 0 && selector.RequiredOriginalCost == nil && selector.VP == nil && len(selector.Resources) == 0 {
		return false
	}

	if len(selector.Tags) > 0 {
		for _, requiredTag := range selector.Tags {
			if !slices.Contains(card.Tags, requiredTag) {
				return false
			}
		}
	}

	if len(selector.CardTypes) > 0 {
		if !slices.Contains(selector.CardTypes, string(card.Type)) {
			return false
		}
	}

	if selector.RequiredOriginalCost != nil {
		if selector.RequiredOriginalCost.Min != nil && card.Cost < *selector.RequiredOriginalCost.Min {
			return false
		}
		if selector.RequiredOriginalCost.Max != nil && card.Cost > *selector.RequiredOriginalCost.Max {
			return false
		}
	}

	if selector.VP != nil {
		hasVP := false
		for _, vp := range card.VPConditions {
			if selector.VP.Min != nil && vp.Amount < *selector.VP.Min {
				continue
			}
			if selector.VP.Max != nil && vp.Amount > *selector.VP.Max {
				continue
			}
			hasVP = true
			break
		}
		if !hasVP {
			return false
		}
	}

	if len(selector.Resources) > 0 {
		if !cardHasAnyResource(card, selector.Resources) {
			return false
		}
	}

	return true
}

// cardHasAnyResource checks if a card has any of the specified resources
// in its resource storage, behavior outputs, or behavior inputs.
func cardHasAnyResource(card *Card, resources []string) bool {
	for _, res := range resources {
		rt := shared.ResourceType(res)
		if card.ResourceStorage != nil && card.ResourceStorage.Type == rt {
			return true
		}
		for _, b := range card.Behaviors {
			for _, o := range b.Outputs {
				if o.ResourceType == rt {
					return true
				}
			}
			for _, i := range b.Inputs {
				if i.ResourceType == rt {
					return true
				}
			}
		}
	}
	return false
}

// MatchesAnySelector checks if a card matches any selector (OR between selectors)
func MatchesAnySelector(card *Card, selectors []shared.Selector) bool {
	if len(selectors) == 0 {
		return false
	}
	for _, sel := range selectors {
		if MatchesSelector(card, sel) {
			return true
		}
	}
	return false
}

// MatchesStandardProjectSelector checks if a project matches a selector
func MatchesStandardProjectSelector(project shared.StandardProject, selector shared.Selector) bool {
	return slices.Contains(selector.StandardProjects, project)
}

// MatchesAnyStandardProjectSelector checks OR between selectors for projects
func MatchesAnyStandardProjectSelector(project shared.StandardProject, selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if MatchesStandardProjectSelector(project, sel) {
			return true
		}
	}
	return false
}

// HasCardSelectors returns true if any selector targets cards
func HasCardSelectors(selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if len(sel.Tags) > 0 || len(sel.CardTypes) > 0 || sel.RequiredOriginalCost != nil || sel.VP != nil || len(sel.Resources) > 0 {
			return true
		}
	}
	return false
}

// HasStandardProjectSelectors returns true if any selector targets standard projects
func HasStandardProjectSelectors(selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if len(sel.StandardProjects) > 0 {
			return true
		}
	}
	return false
}

// HasResourceSelectors returns true if any selector targets resources
func HasResourceSelectors(selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if len(sel.Resources) > 0 {
			return true
		}
	}
	return false
}

// GetResourcesFromSelectors collects all resources from selectors into a single slice
func GetResourcesFromSelectors(selectors []shared.Selector) []string {
	var resources []string
	seen := make(map[string]bool)
	for _, sel := range selectors {
		for _, r := range sel.Resources {
			if !seen[r] {
				seen[r] = true
				resources = append(resources, r)
			}
		}
	}
	return resources
}

// hasPreludeCardType returns true if any selector contains "prelude" in its CardTypes
func hasPreludeCardType(selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if slices.Contains(sel.CardTypes, "prelude") {
			return true
		}
	}
	return false
}

// MatchesResourceSelector checks if a resource type matches a selector
func MatchesResourceSelector(resourceType string, selector shared.Selector) bool {
	return slices.Contains(selector.Resources, resourceType)
}

// MatchesAnyResourceSelector checks OR between selectors for resources
func MatchesAnyResourceSelector(resourceType string, selectors []shared.Selector) bool {
	for _, sel := range selectors {
		if MatchesResourceSelector(resourceType, sel) {
			return true
		}
	}
	return false
}
