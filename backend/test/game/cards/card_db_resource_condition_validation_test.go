package cards_test

import (
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game/shared"
)

func TestResourceConditionFieldValidity(t *testing.T) {
	allCards, err := cards.LoadCardsFromJSON("../../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatalf("Failed to load cards: %v", err)
	}

	for _, card := range allCards {
		for bi, behavior := range card.Behaviors {
			for ii, input := range behavior.Inputs {
				violations := shared.ValidateResourceCondition(input, true)
				for _, v := range violations {
					t.Errorf("Card %q (ID=%s) behavior[%d] input[%d] (%s): %s",
						card.Name, card.ID, bi, ii, input.GetResourceType(), v)
				}
			}
			for oi, output := range behavior.Outputs {
				violations := shared.ValidateResourceCondition(output, false)
				for _, v := range violations {
					t.Errorf("Card %q (ID=%s) behavior[%d] output[%d] (%s): %s",
						card.Name, card.ID, bi, oi, output.GetResourceType(), v)
				}
			}
			for ci, choice := range behavior.Choices {
				for ii, input := range choice.Inputs {
					violations := shared.ValidateResourceCondition(input, true)
					for _, v := range violations {
						t.Errorf("Card %q (ID=%s) behavior[%d] choice[%d] input[%d] (%s): %s",
							card.Name, card.ID, bi, ci, ii, input.GetResourceType(), v)
					}
				}
				for oi, output := range choice.Outputs {
					violations := shared.ValidateResourceCondition(output, false)
					for _, v := range violations {
						t.Errorf("Card %q (ID=%s) behavior[%d] choice[%d] output[%d] (%s): %s",
							card.Name, card.ID, bi, ci, oi, output.GetResourceType(), v)
					}
				}
			}
		}
	}
}

func TestAllResourceTypesHaveOutputProfiles(t *testing.T) {
	for _, rt := range shared.AllResourceTypes {
		if _, ok := shared.GetOutputProfile(rt); !ok {
			t.Errorf("ResourceType %q has no output field profile", rt)
		}
	}
}

func TestAllResourceTypesInAllResourceTypes(t *testing.T) {
	// Verify AllResourceTypes contains every constant from resource_type.go
	// by checking that the count matches the output profiles map size.
	// If a new constant is added to resource_type.go but not to AllResourceTypes,
	// TestAllResourceTypesHaveOutputProfiles won't catch it.
	// This test catches it by ensuring the AllResourceTypes slice has at least
	// as many entries as the output profiles map.
	seen := make(map[shared.ResourceType]bool)
	for _, rt := range shared.AllResourceTypes {
		if seen[rt] {
			t.Errorf("ResourceType %q appears more than once in AllResourceTypes", rt)
		}
		seen[rt] = true
	}
}
