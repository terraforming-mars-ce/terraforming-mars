package game_lifecycle_test

import (
	"context"
	"testing"

	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// setupStartingSelectionGame creates a game in starting_selection phase with all phase data populated.
func setupStartingSelectionGame(t *testing.T, hasPrelude bool) (*game.Game, *turnAction.SelectStartingChoicesAction, string, string) {
	t.Helper()

	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	ctx := context.Background()

	packs := []string{"base-game"}
	if hasPrelude {
		packs = append(packs, "prelude")
	}
	testGame.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  packs,
	})

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "Failed to set active status")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseStartingSelection)
	testutil.AssertNoError(t, err, "Failed to set starting selection phase")

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}
	err = testGame.SetTurnOrder(ctx, playerIDs)
	testutil.AssertNoError(t, err, "Failed to set turn order")

	// Set up corporation phase for all players (use corps without forced first actions)
	corpIDs := []string{"B01", "B02"}
	for i, p := range players {
		testutil.SetPlayerCredits(ctx, p, 100)

		err = testGame.SetSelectCorporationPhase(ctx, p.ID(), &shared.SelectCorporationPhase{
			AvailableCorporations: []string{corpIDs[i], "B03"},
		})
		testutil.AssertNoError(t, err, "Failed to set corporation phase")
	}

	// Set up prelude phase if enabled
	if hasPrelude {
		deck := testGame.Deck()
		for _, p := range players {
			preludeIDs, err := deck.DrawPreludeCards(ctx, 4)
			testutil.AssertNoError(t, err, "Failed to draw prelude cards")

			phase := &shared.SelectPreludeCardsPhase{
				AvailablePreludes: preludeIDs,
				MaxSelectable:     2,
			}
			err = testGame.SetSelectPreludeCardsPhase(ctx, p.ID(), phase)
			testutil.AssertNoError(t, err, "Failed to set prelude phase")
		}
	}

	// Set up starting cards phase for all players
	deck := testGame.Deck()
	for _, p := range players {
		projectCards, err := deck.DrawProjectCards(ctx, 10)
		testutil.AssertNoError(t, err, "Failed to draw project cards")

		phase := &shared.SelectStartingCardsPhase{
			AvailableCards: projectCards,
		}
		err = testGame.SetSelectStartingCardsPhase(ctx, p.ID(), phase)
		testutil.AssertNoError(t, err, "Failed to set starting cards phase")
	}

	action := turnAction.NewSelectStartingChoicesAction(repo, cardRegistry, logger)
	return testGame, action, playerIDs[0], playerIDs[1]
}

func TestSelectStartingChoices_HappyPath_WithPreludes(t *testing.T) {
	testGame, action, playerID1, playerID2 := setupStartingSelectionGame(t, true)
	ctx := context.Background()

	// Override prelude phases with known safe cards (no tile placements)
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testutil.AssertNoError(t, testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &shared.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	}), "set prelude phase for player 1")
	testutil.AssertNoError(t, testGame.SetSelectPreludeCardsPhase(ctx, playerID2, &shared.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	}), "set prelude phase for player 2")

	// Player 1 selects corporation, preludes, and cards in one call
	corpPhase1 := testGame.GetSelectCorporationPhase(playerID1)
	cardsPhase1 := testGame.GetSelectStartingCardsPhase(playerID1)
	corpID1 := corpPhase1.AvailableCorporations[0]

	err := action.Execute(ctx, testGame.ID(), playerID1, corpID1, []string{"P01", "P03"}, cardsPhase1.AvailableCards[:2])
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 1")

	// Player 1 phases should be cleared
	testutil.AssertTrue(t, testGame.GetSelectCorporationPhase(playerID1) == nil, "Player 1 corp phase should be cleared")
	testutil.AssertTrue(t, testGame.GetSelectPreludeCardsPhase(playerID1) == nil, "Player 1 prelude phase should be cleared")
	testutil.AssertTrue(t, testGame.GetSelectStartingCardsPhase(playerID1) == nil, "Player 1 starting cards phase should be cleared")

	// Player 1 should have corporation ID set (for display)
	p1, _ := testGame.GetPlayer(playerID1)
	testutil.AssertEqual(t, corpID1, p1.CorporationID(), "Player 1 should have corporation set")

	// Preludes should NOT be in played cards yet (deferred to init_apply_prelude)
	testutil.AssertFalse(t, p1.PlayedCards().Contains("P01"), "P01 should NOT be in played cards yet")
	testutil.AssertFalse(t, p1.PlayedCards().Contains("P03"), "P03 should NOT be in played cards yet")

	// Deferred choices should be stored
	deferred1 := testGame.GetDeferredStartingChoices(playerID1)
	testutil.AssertTrue(t, deferred1 != nil, "Deferred choices should be stored for player 1")
	testutil.AssertEqual(t, corpID1, deferred1.CorporationID, "Deferred corp should match")

	// Game should still be in starting_selection (player 2 hasn't selected)
	testutil.AssertEqual(t, shared.GamePhaseStartingSelection, testGame.CurrentPhase(), "Game should still be in starting selection phase")

	// Player 2 completes selection
	corpPhase2 := testGame.GetSelectCorporationPhase(playerID2)
	corpID2 := corpPhase2.AvailableCorporations[0]
	err = action.Execute(ctx, testGame.ID(), playerID2, corpID2, []string{"P04", "P07"}, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 2")

	// Game should advance to init_apply_corp (not action phase)
	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Game should advance to init_apply_corp phase")

	// No effects applied yet - waiting for first confirm
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting for confirm")
}

func TestSelectStartingChoices_HappyPath_NoPreludes(t *testing.T) {
	testGame, action, playerID1, playerID2 := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	// Player 1 selects corporation and cards (no preludes)
	corpPhase1 := testGame.GetSelectCorporationPhase(playerID1)
	corpID1 := corpPhase1.AvailableCorporations[0]

	err := action.Execute(ctx, testGame.ID(), playerID1, corpID1, []string{}, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 1")

	// Player 2 selects
	corpPhase2 := testGame.GetSelectCorporationPhase(playerID2)
	corpID2 := corpPhase2.AvailableCorporations[0]
	err = action.Execute(ctx, testGame.ID(), playerID2, corpID2, []string{}, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 2")

	// Should go to init_apply_corp (not directly to action phase)
	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Game should be in init_apply_corp phase")
}

func TestSelectStartingChoices_DeferredStorage(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, true)
	ctx := context.Background()

	// Override with safe preludes
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testutil.AssertNoError(t, testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &shared.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	}), "set prelude phase")

	corpPhase1 := testGame.GetSelectCorporationPhase(playerID1)
	corpID1 := corpPhase1.AvailableCorporations[0]

	err := action.Execute(ctx, testGame.ID(), playerID1, corpID1, []string{"P01", "P03"}, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices")

	// Preludes should NOT be in played cards (deferred)
	p1, _ := testGame.GetPlayer(playerID1)
	testutil.AssertFalse(t, p1.PlayedCards().Contains("P01"), "P01 should NOT be in played cards yet")
	testutil.AssertFalse(t, p1.PlayedCards().Contains("P03"), "P03 should NOT be in played cards yet")

	// Deferred choices should store the selected preludes
	deferred := testGame.GetDeferredStartingChoices(playerID1)
	testutil.AssertTrue(t, deferred != nil, "Deferred choices should be stored")
	testutil.AssertEqual(t, corpID1, deferred.CorporationID, "Deferred corp ID should match")
	testutil.AssertEqual(t, 2, len(deferred.PreludeIDs), "Should have 2 deferred preludes")

	// Unselected preludes should be removed from the game entirely (not in discard pile)
	discardPile := testGame.Deck().DiscardPile()
	discardSet := make(map[string]bool)
	for _, id := range discardPile {
		discardSet[id] = true
	}
	testutil.AssertFalse(t, discardSet["P04"], "P04 should NOT be in discard pile (removed from game)")
	testutil.AssertFalse(t, discardSet["P07"], "P07 should NOT be in discard pile (removed from game)")
}

func TestSelectStartingChoices_Validation_WrongPreludeCount(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, true)
	ctx := context.Background()

	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testutil.AssertNoError(t, testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &shared.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	}), "set prelude phase")

	corpPhase := testGame.GetSelectCorporationPhase(playerID1)
	corpID := corpPhase.AvailableCorporations[0]

	// Selecting only 1 prelude should fail
	err := action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{"P01"}, []string{})
	testutil.AssertError(t, err, "Should fail with wrong number of preludes")

	// Selecting 3 preludes should fail
	err = action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{"P01", "P03", "P04"}, []string{})
	testutil.AssertError(t, err, "Should fail with too many preludes")
}

func TestSelectStartingChoices_Validation_InvalidPreludeID(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, true)
	ctx := context.Background()

	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testutil.AssertNoError(t, testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &shared.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	}), "set prelude phase")

	corpPhase := testGame.GetSelectCorporationPhase(playerID1)
	corpID := corpPhase.AvailableCorporations[0]

	err := action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{"P01", "invalid-prelude"}, []string{})
	testutil.AssertError(t, err, "Should fail with invalid prelude ID")
}

func TestSelectStartingChoices_Validation_WrongPhase(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	// Change to action phase
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")

	corpPhase := testGame.GetSelectCorporationPhase(playerID1)
	corpID := corpPhase.AvailableCorporations[0]

	err := action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{}, []string{})
	testutil.AssertError(t, err, "Should fail when not in starting selection phase")
}

func TestSelectStartingChoices_Validation_InvalidCorporation(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	// Use a corporation ID not in the available list
	err := action.Execute(ctx, testGame.ID(), playerID1, "invalid-corp", []string{}, []string{})
	testutil.AssertError(t, err, "Should fail with invalid corporation ID")

	// Corporation phase should still be present
	testutil.AssertTrue(t, testGame.GetSelectCorporationPhase(playerID1) != nil, "Corporation phase should still be present")
}

func TestSelectStartingChoices_Validation_NoPlayer(t *testing.T) {
	testGame, action, _, _ := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", "B08", []string{}, []string{})
	testutil.AssertError(t, err, "Should fail with nonexistent player")
}

func TestSelectStartingChoices_Validation_AlreadyCompleted(t *testing.T) {
	testGame, action, playerID1, _ := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	corpPhase := testGame.GetSelectCorporationPhase(playerID1)
	corpID := corpPhase.AvailableCorporations[0]

	// Complete selection
	err := action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{}, []string{})
	testutil.AssertNoError(t, err, "First selection should succeed")

	// Try again — should fail (phase already cleared)
	err = action.Execute(ctx, testGame.ID(), playerID1, corpID, []string{}, []string{})
	testutil.AssertError(t, err, "Should fail when selection already completed")
}
