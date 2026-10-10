package dto_test

import (
	"encoding/json"
	"slices"
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/test/testutil"
)

func TestCardDto_PreservesCombinedTagCount(t *testing.T) {
	card := testutil.GetCardByName("Gyropolis")
	encoded, err := json.Marshal(dto.ToCardDto(card))
	testutil.AssertNoError(t, err, "marshal card")
	var result struct {
		Behaviors []struct {
			Outputs []struct {
				Type string `json:"type"`
				Per  struct {
					Tags []string `json:"tags"`
				} `json:"per"`
			} `json:"outputs"`
		} `json:"behaviors"`
	}
	testutil.AssertNoError(t, json.Unmarshal(encoded, &result), "decode API data")
	found := false
	for _, behavior := range result.Behaviors {
		for _, output := range behavior.Outputs {
			if output.Type == "credit-production" {
				if found {
					t.Fatal("production should use one combined count")
				}
				found = true
				testutil.AssertTrue(t, slices.Equal([]string{"earth", "venus"}, output.Per.Tags), "both icons reach frontend")
			}
		}
	}
	if !found {
		t.Fatal("missing production output")
	}
}
