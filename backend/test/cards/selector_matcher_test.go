package cards_test

import (
	"testing"

	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

func TestMatchesSelector_SingleTag(t *testing.T) {
	card := &gamecards.Card{
		ID:   "001",
		Name: "Test Card",
		Type: "automated",
		Tags: []shared.CardTag{"space", "building"},
	}

	selector := shared.Selector{
		Tags: []shared.CardTag{"space"},
	}

	if !gamecards.MatchesSelector(card, selector) {
		t.Error("Expected card with space tag to match selector with space tag")
	}
}

func TestMatchesSelector_MultipleTagsAND(t *testing.T) {
	card := &gamecards.Card{
		ID:   "001",
		Name: "Test Card",
		Type: "automated",
		Tags: []shared.CardTag{"space", "event"},
	}

	selector := shared.Selector{
		Tags: []shared.CardTag{"space", "event"},
	}

	if !gamecards.MatchesSelector(card, selector) {
		t.Error("Expected card with both space and event tags to match selector requiring both")
	}

	cardMissingTag := &gamecards.Card{
		ID:   "002",
		Name: "Test Card 2",
		Type: "automated",
		Tags: []shared.CardTag{"space"},
	}

	if gamecards.MatchesSelector(cardMissingTag, selector) {
		t.Error("Expected card missing event tag to NOT match selector requiring both space and event")
	}
}

func TestMatchesSelector_TagAndCardType(t *testing.T) {
	spaceEvent := &gamecards.Card{
		ID:   "001",
		Name: "Space Event Card",
		Type: "event",
		Tags: []shared.CardTag{"space"},
	}

	spaceAutomated := &gamecards.Card{
		ID:   "002",
		Name: "Space Automated Card",
		Type: "automated",
		Tags: []shared.CardTag{"space"},
	}

	nonSpaceEvent := &gamecards.Card{
		ID:   "003",
		Name: "Non-Space Event Card",
		Type: "event",
		Tags: []shared.CardTag{"building"},
	}

	selector := shared.Selector{
		Tags:      []shared.CardTag{"space"},
		CardTypes: []string{"event"},
	}

	if !gamecards.MatchesSelector(spaceEvent, selector) {
		t.Error("Expected space event to match selector (space + event)")
	}

	if gamecards.MatchesSelector(spaceAutomated, selector) {
		t.Error("Expected space automated to NOT match selector (requires event type)")
	}

	if gamecards.MatchesSelector(nonSpaceEvent, selector) {
		t.Error("Expected non-space event to NOT match selector (requires space tag)")
	}
}

func TestMatchesAnySelector_OrLogic(t *testing.T) {
	plantCard := &gamecards.Card{
		ID:   "001",
		Name: "Plant Card",
		Type: "automated",
		Tags: []shared.CardTag{"plant"},
	}

	microbeCard := &gamecards.Card{
		ID:   "002",
		Name: "Microbe Card",
		Type: "automated",
		Tags: []shared.CardTag{"microbe"},
	}

	spaceCard := &gamecards.Card{
		ID:   "003",
		Name: "Space Card",
		Type: "automated",
		Tags: []shared.CardTag{"space"},
	}

	selectors := []shared.Selector{
		{Tags: []shared.CardTag{"plant"}},
		{Tags: []shared.CardTag{"microbe"}},
		{Tags: []shared.CardTag{"animal"}},
	}

	if !gamecards.MatchesAnySelector(plantCard, selectors) {
		t.Error("Expected plant card to match OR selectors (plant | microbe | animal)")
	}

	if !gamecards.MatchesAnySelector(microbeCard, selectors) {
		t.Error("Expected microbe card to match OR selectors (plant | microbe | animal)")
	}

	if gamecards.MatchesAnySelector(spaceCard, selectors) {
		t.Error("Expected space card to NOT match OR selectors (plant | microbe | animal)")
	}
}

func TestMatchesStandardProjectSelector(t *testing.T) {
	selector := shared.Selector{
		StandardProjects: []shared.StandardProject{
			shared.StandardProjectPowerPlant,
			shared.StandardProjectAquifer,
		},
	}

	if !gamecards.MatchesStandardProjectSelector(shared.StandardProjectPowerPlant, selector) {
		t.Error("Expected power-plant to match selector")
	}

	if !gamecards.MatchesStandardProjectSelector(shared.StandardProjectAquifer, selector) {
		t.Error("Expected aquifer to match selector")
	}

	if gamecards.MatchesStandardProjectSelector(shared.StandardProjectGreenery, selector) {
		t.Error("Expected greenery to NOT match selector")
	}
}

func TestHasCardSelectors(t *testing.T) {
	cardSelectors := []shared.Selector{
		{Tags: []shared.CardTag{"space"}},
	}

	projectSelectors := []shared.Selector{
		{StandardProjects: []shared.StandardProject{shared.StandardProjectPowerPlant}},
	}

	mixedSelectors := []shared.Selector{
		{Tags: []shared.CardTag{"power"}},
		{StandardProjects: []shared.StandardProject{shared.StandardProjectPowerPlant}},
	}

	if !gamecards.HasCardSelectors(cardSelectors) {
		t.Error("Expected tag-based selectors to return true for HasCardSelectors")
	}

	if gamecards.HasCardSelectors(projectSelectors) {
		t.Error("Expected project-only selectors to return false for HasCardSelectors")
	}

	if !gamecards.HasCardSelectors(mixedSelectors) {
		t.Error("Expected mixed selectors (with tags) to return true for HasCardSelectors")
	}
}

func TestGetResourcesFromSelectors(t *testing.T) {
	selectors := []shared.Selector{
		{Resources: []string{"plant", "microbe"}},
		{Resources: []string{"animal", "plant"}},
	}

	resources := gamecards.GetResourcesFromSelectors(selectors)

	if len(resources) != 3 {
		t.Errorf("Expected 3 unique resources, got %d", len(resources))
	}

	expected := map[string]bool{"plant": true, "microbe": true, "animal": true}
	for _, r := range resources {
		if !expected[r] {
			t.Errorf("Unexpected resource: %s", r)
		}
	}
}

func intPtr(i int) *int {
	return &i
}

func TestMatchesSelector_RequiredOriginalCost(t *testing.T) {
	// Card costing 25
	card := &gamecards.Card{
		ID:   "001",
		Name: "Expensive Card",
		Type: "automated",
		Cost: 25,
	}

	// Selector requiring min 20
	selector := shared.Selector{
		RequiredOriginalCost: &shared.MinMaxValue{Min: intPtr(20)},
	}

	if !gamecards.MatchesSelector(card, selector) {
		t.Error("Card costing 25 should match min 20 selector")
	}

	// Card costing 15 should NOT match
	cheapCard := &gamecards.Card{
		ID:   "002",
		Name: "Cheap Card",
		Type: "automated",
		Cost: 15,
	}

	if gamecards.MatchesSelector(cheapCard, selector) {
		t.Error("Card costing 15 should NOT match min 20 selector")
	}

	// Test max constraint
	maxSelector := shared.Selector{
		RequiredOriginalCost: &shared.MinMaxValue{Max: intPtr(20)},
	}

	if gamecards.MatchesSelector(card, maxSelector) {
		t.Error("Card costing 25 should NOT match max 20 selector")
	}

	if !gamecards.MatchesSelector(cheapCard, maxSelector) {
		t.Error("Card costing 15 should match max 20 selector")
	}
}

func TestMatchesSelector_CostWithTags(t *testing.T) {
	// Space event costing 25
	spaceEventExpensive := &gamecards.Card{
		ID:   "001",
		Name: "Expensive Space Event",
		Type: "event",
		Cost: 25,
		Tags: []shared.CardTag{"space"},
	}

	// Space event costing 15
	spaceEventCheap := &gamecards.Card{
		ID:   "002",
		Name: "Cheap Space Event",
		Type: "event",
		Cost: 15,
		Tags: []shared.CardTag{"space"},
	}

	// Non-space event costing 25
	nonSpaceEventExpensive := &gamecards.Card{
		ID:   "003",
		Name: "Expensive Non-Space Event",
		Type: "event",
		Cost: 25,
		Tags: []shared.CardTag{"building"},
	}

	// Selector: space + event + cost 20+
	selector := shared.Selector{
		Tags:                 []shared.CardTag{"space"},
		CardTypes:            []string{"event"},
		RequiredOriginalCost: &shared.MinMaxValue{Min: intPtr(20)},
	}

	if !gamecards.MatchesSelector(spaceEventExpensive, selector) {
		t.Error("Space event costing 25 should match (space + event + cost 20+)")
	}

	if gamecards.MatchesSelector(spaceEventCheap, selector) {
		t.Error("Space event costing 15 should NOT match (requires cost 20+)")
	}

	if gamecards.MatchesSelector(nonSpaceEventExpensive, selector) {
		t.Error("Non-space event costing 25 should NOT match (requires space tag)")
	}
}

func TestHasCardSelectors_WithCost(t *testing.T) {
	// Cost-only selectors should return true for HasCardSelectors
	costSelectors := []shared.Selector{
		{RequiredOriginalCost: &shared.MinMaxValue{Min: intPtr(20)}},
	}

	if !gamecards.HasCardSelectors(costSelectors) {
		t.Error("Cost-based selectors should return true for HasCardSelectors")
	}

	// Mixed cost and project selectors
	mixedSelectors := []shared.Selector{
		{StandardProjects: []shared.StandardProject{shared.StandardProjectPowerPlant}},
		{RequiredOriginalCost: &shared.MinMaxValue{Min: intPtr(20)}},
	}

	if !gamecards.HasCardSelectors(mixedSelectors) {
		t.Error("Mixed selectors with cost should return true for HasCardSelectors")
	}
}

func TestMatchesSelector_VP_NonNegative(t *testing.T) {
	cardWithPositiveVP := &gamecards.Card{
		ID:   "v01",
		Name: "Card With VP",
		Type: "automated",
		VPConditions: []gamecards.VictoryPointCondition{
			{Amount: 1, Condition: "once"},
		},
	}

	cardWithZeroVP := &gamecards.Card{
		ID:   "v02",
		Name: "Card With Zero VP",
		Type: "automated",
		VPConditions: []gamecards.VictoryPointCondition{
			{Amount: 0, Condition: "once"},
		},
	}

	cardWithNegativeVP := &gamecards.Card{
		ID:   "v03",
		Name: "Card With Negative VP",
		Type: "automated",
		VPConditions: []gamecards.VictoryPointCondition{
			{Amount: -1, Condition: "once"},
		},
	}

	cardWithNoVP := &gamecards.Card{
		ID:   "v04",
		Name: "Card Without VP",
		Type: "automated",
	}

	selector := shared.Selector{
		VP: &shared.MinMaxValue{Min: intPtr(0)},
	}

	if !gamecards.MatchesSelector(cardWithPositiveVP, selector) {
		t.Error("Card with VP=1 should match VP min:0 selector")
	}

	if !gamecards.MatchesSelector(cardWithZeroVP, selector) {
		t.Error("Card with VP=0 should match VP min:0 selector")
	}

	if gamecards.MatchesSelector(cardWithNegativeVP, selector) {
		t.Error("Card with VP=-1 should NOT match VP min:0 selector")
	}

	if gamecards.MatchesSelector(cardWithNoVP, selector) {
		t.Error("Card without VP conditions should NOT match VP selector")
	}
}

func TestMatchesSelector_VP_WithMaxConstraint(t *testing.T) {
	cardWithVP3 := &gamecards.Card{
		ID:   "v05",
		Name: "Card With 3 VP",
		Type: "automated",
		VPConditions: []gamecards.VictoryPointCondition{
			{Amount: 3, Condition: "once"},
		},
	}

	selector := shared.Selector{
		VP: &shared.MinMaxValue{Min: intPtr(1), Max: intPtr(2)},
	}

	if gamecards.MatchesSelector(cardWithVP3, selector) {
		t.Error("Card with VP=3 should NOT match VP min:1 max:2 selector")
	}

	selectorMaxOnly := shared.Selector{
		VP: &shared.MinMaxValue{Max: intPtr(2)},
	}

	if gamecards.MatchesSelector(cardWithVP3, selectorMaxOnly) {
		t.Error("Card with VP=3 should NOT match VP max:2 selector")
	}
}

func TestHasCardSelectors_VP(t *testing.T) {
	vpSelectors := []shared.Selector{
		{VP: &shared.MinMaxValue{Min: intPtr(0)}},
	}

	if !gamecards.HasCardSelectors(vpSelectors) {
		t.Error("VP-based selectors should return true for HasCardSelectors")
	}
}

func TestOptimalAerobrakingScenario(t *testing.T) {
	spaceEventCard := &gamecards.Card{
		ID:   "087",
		Name: "Asteroid",
		Type: "event",
		Tags: []shared.CardTag{"space", "event"},
	}

	spaceAutomatedCard := &gamecards.Card{
		ID:   "088",
		Name: "Space Station",
		Type: "active",
		Tags: []shared.CardTag{"space"},
	}

	nonSpaceEventCard := &gamecards.Card{
		ID:   "089",
		Name: "Other Event",
		Type: "event",
		Tags: []shared.CardTag{"building"},
	}

	optimalAerobrakingSelectors := []shared.Selector{
		{
			Tags:      []shared.CardTag{"space"},
			CardTypes: []string{"event"},
		},
	}

	if !gamecards.MatchesAnySelector(spaceEventCard, optimalAerobrakingSelectors) {
		t.Error("Optimal Aerobraking should trigger on space event card")
	}

	if gamecards.MatchesAnySelector(spaceAutomatedCard, optimalAerobrakingSelectors) {
		t.Error("Optimal Aerobraking should NOT trigger on space non-event card")
	}

	if gamecards.MatchesAnySelector(nonSpaceEventCard, optimalAerobrakingSelectors) {
		t.Error("Optimal Aerobraking should NOT trigger on non-space event card")
	}
}
