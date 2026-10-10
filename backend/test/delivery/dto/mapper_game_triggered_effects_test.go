package dto_test

import (
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

// TestToGameDto_TriggeredEffectsNotCleared verifies that calling ToGameDto
// multiple times for different players returns the same triggered effects.
// This is critical for multiplayer: the mapper must not mutate game state.
func TestToGameDto_TriggeredEffectsNotCleared(t *testing.T) {
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()

	testGame.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:   "Test Card",
		PlayerID:   players[0].ID(),
		SourceType: shared.SourceTypePassiveEffect,
	})

	// Build DTOs for each player sequentially (same order as broadcaster loop)
	for _, p := range players {
		gameDto := dto.ToGameDto(testGame, cardRegistry, p.ID())

		if len(gameDto.TriggeredEffects) != 1 {
			t.Fatalf("player %s: expected 1 triggered effect, got %d", p.ID(), len(gameDto.TriggeredEffects))
		}
		if gameDto.TriggeredEffects[0].CardName != "Test Card" {
			t.Fatalf("player %s: expected card name 'Test Card', got '%s'", p.ID(), gameDto.TriggeredEffects[0].CardName)
		}
	}

	// Effects should still be present on the game until explicitly cleared
	effects := testGame.GetTriggeredEffects()
	if len(effects) != 1 {
		t.Fatalf("expected triggered effects to still be present on game, got %d", len(effects))
	}
}

func TestResourceRemoval_DTOIncludesRestrictionAndSelection(t *testing.T) {
	g, _, registry, id, otherID := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	c := testutil.GetCardByName("Comet For Venus")
	card := dto.ToCardDto(c)
	output := card.Behaviors[0].Outputs[1].(dto.BasicResourceConditionDto)
	if output.TargetRestriction == nil || len(output.TargetRestriction.Selectors) != 1 || len(output.TargetRestriction.Selectors[0].Tags) != 1 || output.TargetRestriction.Selectors[0].Tags[0] != dto.CardTag("venus") {
		t.Fatal("missing Venus target restriction")
	}
	p.Selection().SetPendingResourceRemovalSelection(&shared.PendingResourceRemovalSelection{ID: "selection", SourceCardID: c.ID, Source: c.Name, ResourceType: shared.ResourceCredit, Amount: 4, EligiblePlayerIDs: []string{otherID}, MaxAmounts: map[string]int{otherID: 2}})
	state := dto.ToGameDto(g, registry, id)
	pending := state.CurrentPlayer.PendingResourceRemovalSelection
	if pending == nil || pending.ID != "selection" || pending.MaxAmounts[otherID] != 2 || pending.ResourceType != "credit" {
		t.Fatal("pending removal DTO lost selection data")
	}
}

func TestToGameDto_ResourceRemovalTargetsUseViewingPlayerPolicy(t *testing.T) {
	g, _, registry, ownerID, actorID := testutil.SetupTwoPlayerGame(t)
	owner, _ := g.GetPlayer(ownerID)
	actor, _ := g.GetPlayer(actorID)
	owner.Resources().Add(map[shared.ResourceType]int{shared.ResourcePlant: 5, shared.ResourceHeat: 3})
	actor.Resources().Add(map[shared.ResourceType]int{shared.ResourcePlant: 2})
	pets := testutil.GetCardByName("Pets")
	habitats := testutil.GetCardByName("Protected Habitats")
	corp := testutil.GetCardByName("Arklight")
	owner.SetCorporationID(corp.ID)
	owner.Resources().AddToStorage(corp.ID, 2)
	owner.PlayedCards().AddCard(pets.ID, pets.Name, string(pets.Type), nil)
	owner.Resources().AddToStorage(pets.ID, 1)
	owner.Effects().AddEffect(shared.CardEffect{CardID: pets.ID, Behavior: pets.Behaviors[0]})
	owner.Effects().AddEffect(shared.CardEffect{CardID: habitats.ID, Behavior: habitats.Behaviors[0]})
	for _, viewer := range []string{ownerID, actorID} {
		mapped := dto.ToGameDto(g, registry, viewer)
		foundPlant, foundCorp, foundHeat := false, false, false
		for _, target := range mapped.CurrentPlayer.ResourceRemovalTargets {
			if target.CardID == pets.ID {
				t.Fatal("Pets exposed as a legal target")
			}
			if target.PlayerID != ownerID {
				continue
			}
			if target.CardID == corp.ID {
				foundCorp = true
				testutil.AssertEqual(t, 2, target.Amount, "corp storage amount")
			}
			if target.ResourceType == dto.ResourceType("plant") {
				foundPlant = true
			}
			if target.ResourceType == dto.ResourceType("heat") {
				foundHeat = true
			}
		}
		testutil.AssertEqual(t, viewer == ownerID, foundPlant, "owner plants only removable by self")
		testutil.AssertEqual(t, viewer == ownerID, foundCorp, "corporation storage protected")
		testutil.AssertTrue(t, foundHeat, "unrelated resource remains targetable")
	}
}

func TestCardDto_TileAreaMigration(t *testing.T) {
	for _, tc := range []struct{ name, area string }{
		{"Artificial Lake", "land"}, {"Mangrove", "ocean"}, {"Protected Valley", "ocean"}, {"Mohole Area", "ocean"},
	} {
		card := testutil.GetCardByName(tc.name)
		mapped := dto.ToCardDto(card)
		found := false
		for _, behavior := range mapped.Behaviors {
			for _, output := range behavior.Outputs {
				tile, ok := output.(dto.TilePlacementConditionDto)
				if !ok || tile.TileRestrictions == nil {
					continue
				}
				testutil.AssertEqual(t, tc.area, tile.TileRestrictions.Area, "area sent to clients")
				found = true
			}
		}
		testutil.AssertTrue(t, found, "card exposes migrated area")
	}
}
