package cards_test

import (
	"testing"

	"terraforming-mars-backend/internal/game/shared"
)

// TestAutoSelectChoiceIndex_PlantTags3_Below verifies that with <3 plant tags,
// choice 0 (1 plant production) is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_Below(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(shared.ChoicePolicyAutoPlantTags3, 2)
	if idx != 0 {
		t.Fatalf("expected choice 0 with 2 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_PlantTags3_AtThreshold verifies that with exactly 3 plant tags,
// choice 1 (4 plant production) is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_AtThreshold(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(shared.ChoicePolicyAutoPlantTags3, 3)
	if idx != 1 {
		t.Fatalf("expected choice 1 with 3 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_PlantTags3_Above verifies that with >3 plant tags,
// choice 1 is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_Above(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(shared.ChoicePolicyAutoPlantTags3, 5)
	if idx != 1 {
		t.Fatalf("expected choice 1 with 5 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_UnknownPolicy returns -1.
func TestAutoSelectChoiceIndex_UnknownPolicy(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex("unknown-policy", 10)
	if idx != -1 {
		t.Fatalf("expected -1 for unknown policy, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_EmptyPolicy returns -1.
func TestAutoSelectChoiceIndex_EmptyPolicy(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex("", 5)
	if idx != -1 {
		t.Fatalf("expected -1 for empty policy, got %d", idx)
	}
}

// TestFilterChoiceIndicesByPolicy_AutoPlantTags3 verifies the policy returns all indices
// (auto-select is handled externally, not by filtering).
func TestFilterChoiceIndicesByPolicy_AutoPlantTags3(t *testing.T) {
	choices := []shared.Choice{
		{Outputs: []shared.ResourceCondition{{ResourceType: shared.ResourcePlantProduction, Amount: 1}}},
		{Outputs: []shared.ResourceCondition{{ResourceType: shared.ResourcePlantProduction, Amount: 4}}},
	}
	production := shared.Production{}

	indices := shared.FilterChoiceIndicesByPolicy(choices, shared.ChoicePolicyAutoPlantTags3, production)
	if len(indices) != 2 {
		t.Fatalf("expected 2 valid indices for auto-plant-tags-3, got %d", len(indices))
	}
}
