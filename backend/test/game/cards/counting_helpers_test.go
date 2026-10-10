package cards_test

import (
	"testing"

	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
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

func TestTagCounts_ExplicitActionContext(t *testing.T) {
	g, _, registry, id, opponentID := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	p.SetCorporationID("")
	for _, name := range []string{"Research Coordination", "Research Network", "Search For Life"} {
		card := testutil.GetCardByName(name)
		p.PlayedCards().AddCard(card.ID, card.Name, string(card.Type), nil)
	}
	p.AddBonusTags(shared.TagScience, 1)
	for _, tc := range []struct {
		name    string
		tags    []shared.CardTag
		context gamecards.TagCountContext
		want    int
	}{
		{"actual", []shared.CardTag{shared.TagScience}, gamecards.TagCountContext{}, 2},
		{"owner action", []shared.CardTag{shared.TagScience}, gamecards.TagCountContext{ActorID: id}, 4},
		{"opponent action", []shared.CardTag{shared.TagScience}, gamecards.TagCountContext{ActorID: opponentID}, 2},
		{"union", []shared.CardTag{shared.TagScience, shared.TagPlant, shared.TagScience}, gamecards.TagCountContext{ActorID: id}, 4},
		{"new wild excluded", []shared.CardTag{shared.TagScience}, gamecards.TagCountContext{ActorID: id, ExcludeWildCardID: testutil.CardID("Research Coordination")}, 3},
		{"not event", []shared.CardTag{shared.TagEvent}, gamecards.TagCountContext{ActorID: id}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.AssertEqual(t, tc.want, gamecards.CountPlayerTags(p, registry, tc.tags, tc.context), "count")
		})
	}
	testutil.AssertEqual(t, 2, gamecards.CountPlayerTagsByType(p, registry, shared.TagScience), "raw helper never uses wilds")
}

func TestTagRequirements_AllocateSharedWildPool(t *testing.T) {
	one, two, three, zero := 1, 2, 3, 0
	plant, microbe, science := shared.TagPlant, shared.TagMicrobe, shared.TagScience
	for _, tc := range []struct {
		name         string
		counts       gamecards.TagCounts
		requirements []gamecards.Requirement
		valid        bool
	}{
		{"two missing one wild", gamecards.TagCounts{shared.TagWild: 1}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &plant, Min: &one}, {Type: gamecards.RequirementTags, Tag: &microbe, Min: &one}}, false},
		{"two missing two wilds", gamecards.TagCounts{shared.TagWild: 2}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &plant, Min: &one}, {Type: gamecards.RequirementTags, Tag: &microbe, Min: &one}}, true},
		{"science threshold", gamecards.TagCounts{science: 2, shared.TagWild: 1}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &science, Min: &three}}, true},
		{"repeated type uses highest minimum", gamecards.TagCounts{science: 1, shared.TagWild: 2}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &science, Min: &two}, {Type: gamecards.RequirementTags, Tag: &science, Min: &three}}, true},
		{"unassigned wild respects maximum", gamecards.TagCounts{shared.TagWild: 2}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &science, Max: &zero}}, true},
		{"actual tag cannot be hidden", gamecards.TagCounts{science: 1, shared.TagWild: 2}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &science, Max: &zero}}, false},
		{"conflicting bounds", gamecards.TagCounts{shared.TagWild: 2}, []gamecards.Requirement{{Type: gamecards.RequirementTags, Tag: &science, Min: &two}, {Type: gamecards.RequirementTags, Tag: &science, Max: &one}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := gamecards.ValidateTagRequirements(tc.requirements, tc.counts)
			testutil.AssertEqual(t, tc.valid, err == nil, "requirements")
		})
	}
}

func TestWildTags_MilestonesVersusAwards(t *testing.T) {
	g, _, registry, id, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	p.SetCorporationID("")
	wild := testutil.GetCardByName("Research Coordination")
	p.PlayedCards().AddCard(wild.ID, wild.Name, string(wild.Type), nil)
	for _, name := range []string{"builder", "ecologist", "diversifier"} {
		definition, err := testutil.CreateTestMilestoneRegistry().GetByID(name)
		testutil.AssertNoError(t, err, "milestone")
		testutil.AssertEqual(t, 1, gamecards.CalculateMilestoneProgress(definition, p, g.Board(), registry), "one wild contributes once")
	}
	for _, name := range []string{"scientist", "space-baron"} {
		definition, err := testutil.CreateTestAwardRegistry().GetByID(name)
		testutil.AssertNoError(t, err, "award")
		testutil.AssertEqual(t, 0, gamecards.CalculateAwardScore(definition, p, g.Board(), registry), "wild grants no award score")
	}
}
