package behavior_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestColony_PlacementCreatesPendingSelection(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	// Enable colonies expansion
	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, shared.PackColonies)
	testGame.UpdateSettings(context.Background(), settings)

	// Set up colony tiles on the game
	testGame.Colonies().SetStates([]*colony.ColonyState{
		{DefinitionID: "luna", MarkerPosition: 1, PlayerColonies: nil},
		{DefinitionID: "europa", MarkerPosition: 1, PlayerColonies: nil},
	})

	output := shared.NewColonyCondition(shared.ResourceColony, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	selection := p.Selection().GetPendingColonySelection()
	testutil.AssertTrue(t, selection != nil, "Should have a pending colony selection")
	testutil.AssertTrue(t, len(selection.AvailableColonyIDs) >= 2, "Should have at least 2 available colonies")
}

func TestColony_AllowDuplicatePlayerColony(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	// Enable colonies expansion
	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, shared.PackColonies)
	testGame.UpdateSettings(context.Background(), settings)

	// Player already has a colony on luna
	testGame.Colonies().SetStates([]*colony.ColonyState{
		{DefinitionID: "luna", MarkerPosition: 1, PlayerColonies: []string{playerID}},
		{DefinitionID: "europa", MarkerPosition: 1, PlayerColonies: nil},
	})

	output := &shared.ColonyCondition{
		ConditionBase:              shared.ConditionBase{ResourceType: shared.ResourceColony, Amount: 1, Target: "none"},
		AllowDuplicatePlayerColony: true,
	}
	applyOutputs(t, p, testGame, cardRegistry, output)

	selection := p.Selection().GetPendingColonySelection()
	testutil.AssertTrue(t, selection != nil, "Should have a pending colony selection")

	// Verify luna is still in the available list despite player already having a colony there
	foundLuna := false
	for _, id := range selection.AvailableColonyIDs {
		if id == "luna" {
			foundLuna = true
			break
		}
	}
	testutil.AssertTrue(t, foundLuna, "Luna should be available even though player already has a colony there")
}

func TestColony_PlacementNoColoniesAvailable(t *testing.T) {
	testGame, _, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	// Enable colonies expansion
	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, shared.PackColonies)
	testGame.UpdateSettings(context.Background(), settings)

	// Set up all colony tiles as fully colonized (3 colonies each = max)
	testGame.Colonies().SetStates([]*colony.ColonyState{
		{DefinitionID: "luna", MarkerPosition: 1, PlayerColonies: []string{playerID, otherPlayerID, playerID}},
		{DefinitionID: "europa", MarkerPosition: 1, PlayerColonies: []string{playerID, otherPlayerID, otherPlayerID}},
	})

	output := shared.NewColonyCondition(shared.ResourceColony, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	// No colonies available, so no pending selection should be created
	selection := p.Selection().GetPendingColonySelection()
	testutil.AssertTrue(t, selection == nil, "Should have no pending colony selection when all colonies are full")
}
