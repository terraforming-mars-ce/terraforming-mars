package game_lifecycle_test

import (
	"context"
	"testing"

	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// completeAllSelections makes both players select their starting choices and returns the confirm action.
func completeAllSelections(
	t *testing.T,
	testGame *game.Game,
	selectAction *turnAction.SelectStartingChoicesAction,
	playerID1, playerID2 string,
	hasPrelude bool,
) *turnAction.ConfirmInitAdvanceAction {
	t.Helper()
	ctx := context.Background()
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	if hasPrelude {
		safePreludes := []string{"P01", "P03", "P04", "P07"}
		err := testGame.SetSelectPreludeCardsPhase(ctx, playerID1, &shared.SelectPreludeCardsPhase{
			AvailablePreludes: safePreludes,
			MaxSelectable:     2,
		})
		testutil.AssertNoError(t, err, "set prelude phase for player 1")
		err = testGame.SetSelectPreludeCardsPhase(ctx, playerID2, &shared.SelectPreludeCardsPhase{
			AvailablePreludes: safePreludes,
			MaxSelectable:     2,
		})
		testutil.AssertNoError(t, err, "set prelude phase for player 2")
	}

	corpPhase1 := testGame.GetSelectCorporationPhase(playerID1)
	corpID1 := corpPhase1.AvailableCorporations[0]
	preludes1 := []string{}
	if hasPrelude {
		preludes1 = []string{"P01", "P03"}
	}
	err := selectAction.Execute(ctx, testGame.ID(), playerID1, corpID1, preludes1, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 1")

	corpPhase2 := testGame.GetSelectCorporationPhase(playerID2)
	corpID2 := corpPhase2.AvailableCorporations[0]
	preludes2 := []string{}
	if hasPrelude {
		preludes2 = []string{"P04", "P07"}
	}
	err = selectAction.Execute(ctx, testGame.ID(), playerID2, corpID2, preludes2, []string{})
	testutil.AssertNoError(t, err, "Failed to select starting choices for player 2")

	repo, _ := testutil.CreateTestGameWithPlayers(t, 0, testutil.NewMockBroadcaster())
	_ = repo
	return turnAction.NewConfirmInitAdvanceAction(
		gameRepoWithGame(t, testGame),
		cardRegistry,
		nil,
		nil,
		logger,
	)
}

// gameRepoWithGame creates a repo and inserts the given game.
func gameRepoWithGame(t *testing.T, g *game.Game) game.GameRepository {
	t.Helper()
	repo := testutil.NewTestGameRepository(t)
	err := repo.Create(context.Background(), g)
	testutil.AssertNoError(t, err, "Failed to create game in repo")
	return repo
}

func TestInitPhase_CorpAppliedOneAtATime(t *testing.T) {
	testGame, selectAction, playerID1, playerID2 := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	confirmAction := completeAllSelections(t, testGame, selectAction, playerID1, playerID2, false)

	// After both players select, game should be in init_apply_corp with NO effects applied yet
	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Should be in init_apply_corp")
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting for confirm")

	// First confirm: apply player 1's corp
	err := confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Should apply player 1 corp")

	p1, _ := testGame.GetPlayer(playerID1)
	testutil.AssertTrue(t, p1.HasCorporation(), "Player 1 should have corp after confirm")
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting after corp applied")

	// Second confirm: advance past player 1 to player 2
	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Should advance to player 2")
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting for player 2 corp")

	// Third confirm: apply player 2's corp
	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Should apply player 2 corp")

	p2, _ := testGame.GetPlayer(playerID2)
	testutil.AssertTrue(t, p2.HasCorporation(), "Player 2 should have corp after confirm")

	// Fourth confirm: advance past player 2 → action phase (no prelude)
	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Should advance to action phase")

	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Should be in action phase after all corps (no prelude)")
	testutil.AssertTrue(t, testGame.CurrentTurn() != nil, "Current turn should be set")
}

func TestInitPhase_CorpThenPrelude(t *testing.T) {
	testGame, selectAction, playerID1, playerID2 := setupStartingSelectionGame(t, true)
	ctx := context.Background()

	confirmAction := completeAllSelections(t, testGame, selectAction, playerID1, playerID2, true)

	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Should be in init_apply_corp")

	// Apply + advance through both corps (4 confirms: apply p1, advance, apply p2, advance)
	err := confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Apply player 1 corp")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Advance to player 2")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Apply player 2 corp")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Advance past corp phase")

	// Should now be in init_apply_prelude
	testutil.AssertEqual(t, shared.GamePhaseInitApplyPrelude, testGame.CurrentPhase(), "Should be in init_apply_prelude")
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting for confirm in prelude phase")

	// Apply + advance through both preludes
	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Apply player 1 preludes")

	p1, _ := testGame.GetPlayer(playerID1)
	testutil.AssertTrue(t, p1.PlayedCards().Contains("P01"), "P01 should be in played cards after prelude apply")
	testutil.AssertTrue(t, p1.PlayedCards().Contains("P03"), "P03 should be in played cards after prelude apply")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Advance to player 2 preludes")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Apply player 2 preludes")

	p2, _ := testGame.GetPlayer(playerID2)
	testutil.AssertTrue(t, p2.PlayedCards().Contains("P04"), "P04 should be in played cards")
	testutil.AssertTrue(t, p2.PlayedCards().Contains("P07"), "P07 should be in played cards")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Advance to action phase")

	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Should be in action phase")
	testutil.AssertTrue(t, testGame.CurrentTurn() != nil, "Current turn should be set")
}

func TestInitPhase_PreludeSkippedWhenDisabled(t *testing.T) {
	testGame, selectAction, playerID1, playerID2 := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	confirmAction := completeAllSelections(t, testGame, selectAction, playerID1, playerID2, false)

	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Should be in init_apply_corp")

	// Apply + advance through both corps (4 confirms)
	err := confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Apply player 1 corp")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Advance to player 2")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Apply player 2 corp")

	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Advance past corp phase")

	// Should skip prelude phase entirely and go to action
	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Should skip to action phase when no prelude")
}

func TestInitPhase_RejectConfirmWrongPhase(t *testing.T) {
	testGame, _, playerID1, _ := setupStartingSelectionGame(t, false)
	ctx := context.Background()
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	confirmAction := turnAction.NewConfirmInitAdvanceAction(
		gameRepoWithGame(t, testGame),
		cardRegistry,
		nil,
		nil,
		logger,
	)

	// Game is still in starting_selection
	err := confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertError(t, err, "Should reject confirm when not in init phase")
}

func TestInitPhase_RejectConfirmWhenNotWaiting(t *testing.T) {
	testGame, selectAction, playerID1, playerID2 := setupStartingSelectionGame(t, false)
	ctx := context.Background()

	confirmAction := completeAllSelections(t, testGame, selectAction, playerID1, playerID2, false)

	// Game is now in init_apply_corp and waiting for confirm
	testutil.AssertTrue(t, testGame.InitPhaseWaitingForConfirm(), "Should be waiting")

	// Manually clear waiting flag
	testutil.AssertNoError(t, testGame.SetInitPhaseWaitingForConfirm(ctx, false), "clear waiting flag")

	err := confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertError(t, err, "Should reject confirm when not waiting")
}

func TestInitPhase_LastPlayerForcedTilePlacement(t *testing.T) {
	// Set up game where player 2 (last in turn order) has Tharsis Republic (B08)
	// which requires a forced city tile placement
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()
	ctx := context.Background()

	testGame.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base-game"},
	})

	err := testGame.UpdateStatus(ctx, shared.GameStatusActive)
	testutil.AssertNoError(t, err, "Failed to set active status")
	err = testGame.UpdatePhase(ctx, shared.GamePhaseStartingSelection)
	testutil.AssertNoError(t, err, "Failed to set starting selection phase")

	players := testGame.GetAllPlayers()
	playerID1, playerID2 := players[0].ID(), players[1].ID()
	err = testGame.SetTurnOrder(ctx, []string{playerID1, playerID2})
	testutil.AssertNoError(t, err, "Failed to set turn order")

	for _, p := range players {
		testutil.SetPlayerCredits(ctx, p, 100)
	}

	// Player 1 gets a corp without forced first action (B01)
	err = testGame.SetSelectCorporationPhase(ctx, playerID1, &shared.SelectCorporationPhase{
		AvailableCorporations: []string{"B01"},
	})
	testutil.AssertNoError(t, err, "Failed to set corp phase for player 1")

	// Player 2 (last) gets Tharsis Republic (B08) which has a forced city placement
	err = testGame.SetSelectCorporationPhase(ctx, playerID2, &shared.SelectCorporationPhase{
		AvailableCorporations: []string{"B08"},
	})
	testutil.AssertNoError(t, err, "Failed to set corp phase for player 2")

	deck := testGame.Deck()
	for _, p := range players {
		projectCards, drawErr := deck.DrawProjectCards(ctx, 10)
		testutil.AssertNoError(t, drawErr, "Failed to draw project cards")
		err = testGame.SetSelectStartingCardsPhase(ctx, p.ID(), &shared.SelectStartingCardsPhase{
			AvailableCards: projectCards,
		})
		testutil.AssertNoError(t, err, "Failed to set starting cards phase")
	}

	selectAction := turnAction.NewSelectStartingChoicesAction(repo, cardRegistry, nil, logger)

	// Both players select their starting choices
	err = selectAction.Execute(ctx, testGame.ID(), playerID1, "B01", []string{}, []string{})
	testutil.AssertNoError(t, err, "Player 1 selection")

	err = selectAction.Execute(ctx, testGame.ID(), playerID2, "B08", []string{}, []string{})
	testutil.AssertNoError(t, err, "Player 2 selection")

	testutil.AssertEqual(t, shared.GamePhaseInitApplyCorp, testGame.CurrentPhase(), "Should be in init_apply_corp")

	confirmAction := turnAction.NewConfirmInitAdvanceAction(
		gameRepoWithGame(t, testGame),
		cardRegistry,
		nil,
		nil,
		logger,
	)

	// Confirm 1: apply player 1's corp (no forced action)
	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Apply player 1 corp")

	// Confirm 2: advance to player 2
	err = confirmAction.Execute(ctx, testGame.ID(), playerID1)
	testutil.AssertNoError(t, err, "Advance to player 2")

	// Confirm 3: apply player 2's corp (Tharsis Republic - has forced city placement)
	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Apply player 2 corp (Tharsis Republic)")

	// Player 2 should now have a pending tile selection for city placement
	pendingTile := testGame.GetPendingTileSelection(playerID2)
	testutil.AssertTrue(t, pendingTile != nil, "Player 2 should have pending tile selection after Tharsis Republic corp")
	testutil.AssertEqual(t, "city", pendingTile.TileType, "Pending tile should be a city")
	testutil.AssertTrue(t, len(pendingTile.AvailableHexes) > 0, "Should have available hexes for city placement")

	// Confirm should be blocked while tile is pending
	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertError(t, err, "Should reject confirm while tile is pending")

	// Place the city tile
	selectTile := tileAction.NewSelectTileAction(
		gameRepoWithGame(t, testGame),
		cardRegistry,
		stateRepo,
		logger,
	)
	_, err = selectTile.Execute(ctx, testGame.ID(), playerID2, pendingTile.AvailableHexes[0])
	testutil.AssertNoError(t, err, "Player 2 should be able to place city tile")

	// After placing, pending tile should be cleared
	testutil.AssertTrue(t, testGame.GetPendingTileSelection(playerID2) == nil, "Pending tile should be cleared after placement")

	// Confirm 4: advance past player 2 → action phase
	err = confirmAction.Execute(ctx, testGame.ID(), playerID2)
	testutil.AssertNoError(t, err, "Should advance to action phase after tile placed")

	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Should be in action phase")
}
