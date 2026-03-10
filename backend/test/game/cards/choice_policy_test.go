package cards_test

import (
	"testing"

	"terraforming-mars-backend/internal/game/shared"
)

func intPtr(v int) *int { return &v }

func plantTagAutoPolicy() *shared.ChoicePolicy {
	tag := shared.TagPlant
	return &shared.ChoicePolicy{
		Type:    shared.ChoicePolicyTypeAuto,
		Default: intPtr(0),
		Select: &shared.ChoicePolicySelect{
			Option:       1,
			MinMax:       shared.MinMax{Min: intPtr(3)},
			ResourceType: "tag",
			Tag:          &tag,
		},
	}
}

// TestAutoSelectChoiceIndex_PlantTags3_Below verifies that with <3 plant tags,
// the default choice 0 (1 plant production) is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_Below(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(plantTagAutoPolicy(), 2)
	if idx != 0 {
		t.Fatalf("expected choice 0 with 2 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_PlantTags3_AtThreshold verifies that with exactly 3 plant tags,
// choice 1 (4 plant production) is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_AtThreshold(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(plantTagAutoPolicy(), 3)
	if idx != 1 {
		t.Fatalf("expected choice 1 with 3 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_PlantTags3_Above verifies that with >3 plant tags,
// choice 1 is auto-selected.
func TestAutoSelectChoiceIndex_PlantTags3_Above(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(plantTagAutoPolicy(), 5)
	if idx != 1 {
		t.Fatalf("expected choice 1 with 5 plant tags, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_NilPolicy returns -1.
func TestAutoSelectChoiceIndex_NilPolicy(t *testing.T) {
	idx := shared.AutoSelectChoiceIndex(nil, 10)
	if idx != -1 {
		t.Fatalf("expected -1 for nil policy, got %d", idx)
	}
}

// TestAutoSelectChoiceIndex_LowestPolicy returns -1 (not an auto policy).
func TestAutoSelectChoiceIndex_LowestPolicy(t *testing.T) {
	policy := &shared.ChoicePolicy{Type: shared.ChoicePolicyTypeLowest}
	idx := shared.AutoSelectChoiceIndex(policy, 5)
	if idx != -1 {
		t.Fatalf("expected -1 for lowest policy, got %d", idx)
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

	indices := shared.FilterChoiceIndicesByPolicy(choices, plantTagAutoPolicy(), production)
	if len(indices) != 2 {
		t.Fatalf("expected 2 valid indices for auto policy, got %d", len(indices))
	}
}
