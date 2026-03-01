package game_lifecycle_test

import (
	"context"
	"testing"

	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/test/testutil"
)

// setupPreludeGame creates a game in prelude selection phase with prelude cards dealt.
// Players already have corporations set (simulating Corp → Prelude flow).
func setupPreludeGame(t *testing.T) (*game.Game, game.GameRepository, *turnAction.SelectStartingCardsAction, *turnAction.SelectPreludeCardsAction, string, string) {
	t.Helper()

	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	ctx := context.Background()
	testGame.UpdateSettings(ctx, game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base-game", "prelude"},
	})

	err := testGame.UpdateStatus(ctx, game.GameStatusActive)
	testutil.AssertNoError(t, err, "Failed to set active status")
	err = testGame.UpdatePhase(ctx, game.GamePhasePreludeSelection)
	testutil.AssertNoError(t, err, "Failed to set prelude selection phase")

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}
	err = testGame.SetTurnOrder(ctx, playerIDs)
	testutil.AssertNoError(t, err, "Failed to set turn order")

	// Set corporations on players (Corp selection already done before prelude phase)
	corpIDs := []string{"B08", "B06"} // Tharsis Republic, Mining Guild
	for i, p := range players {
		p.SetCorporationID(corpIDs[i])
		testutil.SetPlayerCredits(ctx, p, 100)
	}

	// Deal prelude cards
	deck := testGame.Deck()
	for _, p := range players {
		preludeIDs, err := deck.DrawPreludeCards(ctx, 4)
		testutil.AssertNoError(t, err, "Failed to draw prelude cards")

		phase := &player.SelectPreludeCardsPhase{
			AvailablePreludes: preludeIDs,
			MaxSelectable:     2,
		}
		err = testGame.SetSelectPreludeCardsPhase(ctx, p.ID(), phase)
		testutil.AssertNoError(t, err, "Failed to set prelude phase")
	}

	selectStartingCardsAction := turnAction.NewSelectStartingCardsAction(repo, cardRegistry, logger)
	selectPreludeCardsAction := turnAction.NewSelectPreludeCardsAction(repo, cardRegistry, logger)

	return testGame, repo, selectStartingCardsAction, selectPreludeCardsAction, playerIDs[0], playerIDs[1]
}

func completePreludeSelection(t *testing.T, g *game.Game, action *turnAction.SelectPreludeCardsAction, playerID string) {
	t.Helper()
	ctx := context.Background()

	// Override with safe preludes (no tile placements)
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	g.SetSelectPreludeCardsPhase(ctx, playerID, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})

	err := action.Execute(ctx, g.ID(), playerID, []string{"P01", "P03"})
	testutil.AssertNoError(t, err, "Failed to complete prelude selection")
}

func TestSelectPreludeCards_TransitionToStartingCards(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, playerID2 := setupPreludeGame(t)

	// Both players complete prelude selection
	completePreludeSelection(t, testGame, selectPreludeCardsAction, playerID1)
	completePreludeSelection(t, testGame, selectPreludeCardsAction, playerID2)

	// Game should transition to starting card selection phase
	testutil.AssertEqual(t, game.GamePhaseStartingCardSelection, testGame.CurrentPhase(), "Game should be in starting card selection phase")

	// Both players should have starting cards (project cards only, no corporations)
	phase1 := testGame.GetSelectStartingCardsPhase(playerID1)
	phase2 := testGame.GetSelectStartingCardsPhase(playerID2)
	testutil.AssertTrue(t, phase1 != nil, "Player 1 should have starting cards phase")
	testutil.AssertTrue(t, phase2 != nil, "Player 2 should have starting cards phase")
	testutil.AssertEqual(t, 10, len(phase1.AvailableCards), "Player 1 should have 10 project cards")
}

func TestSelectPreludeCards_NoPreludePackSkipsPhase(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	testGame.UpdateSettings(ctx, game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base-game"},
	})

	err := testGame.UpdateStatus(ctx, game.GameStatusActive)
	testutil.AssertNoError(t, err, "Failed to set active status")
	err = testGame.UpdatePhase(ctx, game.GamePhaseStartingCardSelection)
	testutil.AssertNoError(t, err, "Failed to set phase")

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}
	err = testGame.SetTurnOrder(ctx, playerIDs)
	testutil.AssertNoError(t, err, "Failed to set turn order")

	deck := testGame.Deck()
	for _, p := range players {
		projectCards, err := deck.DrawProjectCards(ctx, 10)
		testutil.AssertNoError(t, err, "Failed to draw project cards")

		phase := &player.SelectStartingCardsPhase{
			AvailableCards: projectCards,
		}
		err = testGame.SetSelectStartingCardsPhase(ctx, p.ID(), phase)
		testutil.AssertNoError(t, err, "Failed to set starting cards phase")

		// Players already have corporations (from corp selection phase)
		p.SetCorporationID("B08")
		testutil.SetPlayerCredits(ctx, p, 100)
	}

	selectStartingCardsAction := turnAction.NewSelectStartingCardsAction(repo, cardRegistry, logger)

	// Both players complete starting card selection (no corporationID needed)
	for _, pid := range playerIDs {
		err := selectStartingCardsAction.Execute(ctx, testGame.ID(), pid, []string{})
		testutil.AssertNoError(t, err, "Failed to complete starting card selection")
	}

	// Should go straight to action phase
	testutil.AssertEqual(t, game.GamePhaseAction, testGame.CurrentPhase(), "Game should be in action phase without prelude pack")
}

func TestSelectPreludeCards_HappyPath(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, _ := setupPreludeGame(t)

	// Player 1 selects their preludes
	phase1 := testGame.GetSelectPreludeCardsPhase(playerID1)
	selected := []string{phase1.AvailablePreludes[0], phase1.AvailablePreludes[1]}
	unselected := []string{phase1.AvailablePreludes[2], phase1.AvailablePreludes[3]}

	ctx := context.Background()
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, selected)
	testutil.AssertNoError(t, err, "Failed to select prelude cards")

	// Player 1 should have their phase cleared
	testutil.AssertTrue(t, testGame.GetSelectPreludeCardsPhase(playerID1) == nil, "Player 1 prelude phase should be cleared")

	// Selected preludes should be in played cards
	p1, _ := testGame.GetPlayer(playerID1)
	for _, id := range selected {
		testutil.AssertTrue(t, p1.PlayedCards().Contains(id), "Selected prelude should be in played cards")
	}

	// Unselected preludes should be in discard pile
	discardPile := testGame.Deck().DiscardPile()
	discardSet := make(map[string]bool)
	for _, id := range discardPile {
		discardSet[id] = true
	}
	for _, id := range unselected {
		testutil.AssertTrue(t, discardSet[id], "Unselected prelude should be in discard pile")
	}

	// Game should still be in prelude phase (player 2 hasn't selected)
	testutil.AssertEqual(t, game.GamePhasePreludeSelection, testGame.CurrentPhase(), "Game should still be in prelude selection phase")
}

func TestSelectPreludeCards_AllPlayersComplete_AdvancesToStartingCards(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, playerID2 := setupPreludeGame(t)
	ctx := context.Background()

	// Override prelude phases with known safe cards (no tile placements)
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})
	testGame.SetSelectPreludeCardsPhase(ctx, playerID2, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})

	// Both players select preludes
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{"P01", "P03"})
	testutil.AssertNoError(t, err, "Failed to select preludes for player 1")

	err = selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID2, []string{"P04", "P07"})
	testutil.AssertNoError(t, err, "Failed to select preludes for player 2")

	// Game should now be in starting card selection phase
	testutil.AssertEqual(t, game.GamePhaseStartingCardSelection, testGame.CurrentPhase(), "Game should advance to starting card selection phase")

	// Both players should have starting cards dealt (project cards only)
	phase1 := testGame.GetSelectStartingCardsPhase(playerID1)
	phase2 := testGame.GetSelectStartingCardsPhase(playerID2)
	testutil.AssertTrue(t, phase1 != nil, "Player 1 should have starting cards phase")
	testutil.AssertTrue(t, phase2 != nil, "Player 2 should have starting cards phase")
}

func TestSelectPreludeCards_FullFlow_PreludeToAction(t *testing.T) {
	testGame, _, selectStartingCardsAction, selectPreludeCardsAction, playerID1, playerID2 := setupPreludeGame(t)
	ctx := context.Background()

	// Override prelude phases with known safe cards
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})
	testGame.SetSelectPreludeCardsPhase(ctx, playerID2, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})

	// Step 1: Both players complete prelude selection
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{"P01", "P03"})
	testutil.AssertNoError(t, err, "Failed to select preludes for player 1")
	err = selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID2, []string{"P04", "P07"})
	testutil.AssertNoError(t, err, "Failed to select preludes for player 2")

	testutil.AssertEqual(t, game.GamePhaseStartingCardSelection, testGame.CurrentPhase(), "Should be in starting card selection after preludes")

	// Step 2: Both players complete starting card selection (no corporationID needed)
	err = selectStartingCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{})
	testutil.AssertNoError(t, err, "Failed to complete starting cards for player 1")

	err = selectStartingCardsAction.Execute(ctx, testGame.ID(), playerID2, []string{})
	testutil.AssertNoError(t, err, "Failed to complete starting cards for player 2")

	// Game should now be in action phase
	testutil.AssertEqual(t, game.GamePhaseAction, testGame.CurrentPhase(), "Game should advance to action phase")

	// Current turn should be set
	currentTurn := testGame.CurrentTurn()
	testutil.AssertTrue(t, currentTurn != nil, "Current turn should be set")
	testutil.AssertEqual(t, playerID1, currentTurn.PlayerID(), "First player should have the turn")
}

func TestSelectPreludeCards_Validation_WrongCount(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, _ := setupPreludeGame(t)
	ctx := context.Background()

	phase := testGame.GetSelectPreludeCardsPhase(playerID1)

	// Selecting only 1 should fail
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, phase.AvailablePreludes[:1])
	testutil.AssertError(t, err, "Should fail with wrong number of preludes")

	// Selecting 3 should fail
	err = selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, append(phase.AvailablePreludes[:2], phase.AvailablePreludes[2]))
	testutil.AssertError(t, err, "Should fail with too many preludes")
}

func TestSelectPreludeCards_Validation_InvalidID(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, _ := setupPreludeGame(t)
	ctx := context.Background()

	phase := testGame.GetSelectPreludeCardsPhase(playerID1)

	// Selecting an invalid ID should fail
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{phase.AvailablePreludes[0], "invalid-prelude-id"})
	testutil.AssertError(t, err, "Should fail with invalid prelude ID")
}

func TestSelectPreludeCards_Validation_WrongPhase(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, playerID1, playerID2 := setupPreludeGame(t)
	ctx := context.Background()

	// Override to starting card selection phase
	testGame.UpdatePhase(ctx, game.GamePhaseStartingCardSelection)

	// Try to select preludes in wrong phase
	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{"P01", "P02"})
	testutil.AssertError(t, err, "Should fail when not in prelude selection phase")

	// Switch back to prelude phase
	testGame.UpdatePhase(ctx, game.GamePhasePreludeSelection)

	// Now select preludes for player 1 (with safe cards)
	safePreludes := []string{"P01", "P03", "P04", "P07"}
	testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})
	testGame.SetSelectPreludeCardsPhase(ctx, playerID2, &player.SelectPreludeCardsPhase{
		AvailablePreludes: safePreludes,
		MaxSelectable:     2,
	})
	err = selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{"P01", "P03"})
	testutil.AssertNoError(t, err, "Should succeed in prelude phase")

	// Trying again should fail (phase already cleared)
	err = selectPreludeCardsAction.Execute(ctx, testGame.ID(), playerID1, []string{"P01", "P03"})
	testutil.AssertError(t, err, "Should fail when prelude phase already cleared")
}

func TestSelectPreludeCards_Validation_NoPlayer(t *testing.T) {
	testGame, _, _, selectPreludeCardsAction, _, _ := setupPreludeGame(t)
	ctx := context.Background()

	err := selectPreludeCardsAction.Execute(ctx, testGame.ID(), "nonexistent-player", []string{"P01", "P02"})
	testutil.AssertError(t, err, "Should fail with nonexistent player")
}
