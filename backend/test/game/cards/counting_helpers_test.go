package cards_test

import (
	"testing"

	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestCountPlayerTagsByType_ExcludesEventCards verifies that event cards'
// tags are not counted toward non-event tag totals.
func TestCountPlayerTagsByType_ExcludesEventCards(t *testing.T) {
	g, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)

	// Add a regular science card
	p.PlayedCards().AddCard("regular-science", "Research", "automated", []string{"science"})
	// Add an event card with science tag (should NOT count for science)
	p.PlayedCards().AddCard("event-science", "Giant Ice Asteroid", "event", []string{"science", "space", "event"})

	testCards := []gamecards.Card{
		{ID: "regular-science", Name: "Research", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}},
		{ID: "event-science", Name: "Giant Ice Asteroid", Type: gamecards.CardTypeEvent, Tags: []shared.CardTag{shared.TagScience, shared.TagSpace, shared.TagEvent}},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(testCards)

	scienceCount := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagScience)
	// Should be 1: regular card only (event card's science tag excluded)
	if scienceCount != 1 {
		t.Fatalf("expected 1 science tag (event excluded), got %d", scienceCount)
	}

	spaceCount := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagSpace)
	// Should be 0: the only space tag is on the event card
	if spaceCount != 0 {
		t.Fatalf("expected 0 space tags (event excluded), got %d", spaceCount)
	}
}

// TestCountPlayerTagsByType_WildTagCountsForAny verifies that wild tags
// count toward any tag type.
func TestCountPlayerTagsByType_WildTagCountsForAny(t *testing.T) {
	g, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(playerID)

	testCards := []gamecards.Card{
		{ID: "wild-card", Name: "Wild Card", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagWild}},
		{ID: "science-card", Name: "Lab", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagScience}},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(testCards)

	// Measure baseline from corporation (Tharsis Republic has building tag)
	baselineScience := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagScience)
	baselineBuilding := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagBuilding)

	// Now add a wild card and a science card
	p.PlayedCards().AddCard("wild-card", "Wild Card", "automated", []string{"wild"})
	p.PlayedCards().AddCard("science-card", "Lab", "automated", []string{"science"})

	// Science should increase by 2: 1 direct + 1 wild
	scienceCount := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagScience)
	scienceDelta := scienceCount - baselineScience
	if scienceDelta != 2 {
		t.Fatalf("expected science tags to increase by 2 (1 + 1 wild), increased by %d", scienceDelta)
	}

	// Building should increase by 1: just the wild
	buildingCount := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagBuilding)
	buildingDelta := buildingCount - baselineBuilding
	if buildingDelta != 1 {
		t.Fatalf("expected building tags to increase by 1 (wild), increased by %d", buildingDelta)
	}
}
