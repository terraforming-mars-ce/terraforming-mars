package cards_test

import (
	"testing"

	"terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

func TestNoDuplicatePlainAutoBehaviorDescriptions(t *testing.T) {
	allCards, err := cards.LoadCardsFromJSON("../../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatalf("Failed to load cards: %v", err)
	}

	for _, card := range allCards {
		type behaviorScope struct {
			productionBox string
			description   string
		}
		plainAutoCounts := map[behaviorScope]int{}
		for _, behavior := range card.Behaviors {
			if isPlainAuto(behavior) {
				plainAutoCounts[behaviorScope{behavior.ProductionBox, behavior.Description}]++
			}
		}
		for scope, count := range plainAutoCounts {
			if count > 1 {
				t.Errorf("Card %s (%s) has %d plain auto behaviors for production box %q and description %q; merge behaviors with the same description and copy semantics", card.ID, card.Name, count, scope.productionBox, scope.description)
			}
		}
	}
}

func TestAllBehaviorsHaveDescriptions(t *testing.T) {
	allCards, err := cards.LoadCardsFromJSON("../../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatalf("Failed to load cards: %v", err)
	}

	for _, card := range allCards {
		for i, behavior := range card.Behaviors {
			if behavior.Description == "" && behavior.Group == "" {
				t.Errorf("Card %s (%s) behavior[%d] is missing a description",
					card.ID, card.Name, i)
			}
		}

		for i, vp := range card.VPConditions {
			if vp.Description == "" {
				t.Errorf("Card %s (%s) vpConditions[%d] is missing a description",
					card.ID, card.Name, i)
			}
		}

		if card.ResourceStorage != nil && card.ResourceStorage.Description == "" {
			t.Errorf("Card %s (%s) resourceStorage is missing a description",
				card.ID, card.Name)
		}
	}
}

// isPlainAuto returns true if the behavior has exactly one trigger
// of type "auto" with no condition attached.
func isPlainAuto(b shared.CardBehavior) bool {
	if len(b.Triggers) != 1 {
		return false
	}
	trigger := b.Triggers[0]
	return trigger.Type == shared.TriggerTypeAuto && trigger.Condition == nil
}
