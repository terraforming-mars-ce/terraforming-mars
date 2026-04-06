package cards_test

import (
	"strings"
	"testing"

	"terraforming-mars-backend/internal/cards"
)

func TestNoStealTargetsInInputs(t *testing.T) {
	allCards, err := cards.LoadCardsFromJSON("../../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatalf("Failed to load cards: %v", err)
	}

	for _, card := range allCards {
		for bi, behavior := range card.Behaviors {
			for _, input := range behavior.Inputs {
				if strings.HasPrefix(input.GetTarget(), "steal-") {
					t.Errorf("Card %q behavior[%d] has steal target %q in inputs; steal targets belong in outputs",
						card.Name, bi, input.GetTarget())
				}
			}
		}
	}
}
