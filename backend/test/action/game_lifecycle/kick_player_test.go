package game_lifecycle_test

import (
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func newKickAction(repo game.GameRepository, finalScoring *gameaction.FinalScoringAction) *connection.KickPlayerAction {
	return connection.NewKickPlayerAction(repo, nil, finalScoring, testutil.TestLogger())
}

// ============================================================================
// Green Path
// ============================================================================

func TestKickPlayer_ActiveGame_NotTheirTurn(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	// Find a player that is NOT the current turn player and not the host
	currentTurnID := g.CurrentTurn().PlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID && id != currentTurnID {
			targetID = id
			break
		}
	}

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick non-active player")

	p, _ := g.GetPlayer(targetID)
	testutil.AssertTrue(t, p.HasExited(), "Player should be exited")
	testutil.AssertFalse(t, p.IsConnected(), "Player should be disconnected")
	testutil.AssertTrue(t, p.HasPassed(), "Player should be passed")
	testutil.AssertEqual(t, 3, len(g.GetAllPlayers()), "Player should still be in game")
}

func TestKickPlayer_MovesToBottomOfTurnOrder(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	// Kick the first player in turn order (if not host, else second)
	targetID := playerIDs[0]
	if targetID == hostID {
		targetID = playerIDs[1]
	}

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	turnOrder := g.TurnOrder()
	lastID := turnOrder[len(turnOrder)-1]
	testutil.AssertEqual(t, targetID, lastID, "Kicked player should be at end of turn order")
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestKickPlayer_TheirTurn_AdvancesToNext(t *testing.T) {
	g, repo, _, _ := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	currentTurnID := g.CurrentTurn().PlayerID()

	// If it's the host's turn, we can't kick the host (can't kick self)
	// So set the turn to a non-host player
	if currentTurnID == hostID {
		for _, p := range g.GetAllPlayers() {
			if p.ID() != hostID {
				_ = g.SetCurrentTurn(ctx, p.ID(), 2)
				currentTurnID = p.ID()
				break
			}
		}
	}

	err := kick.Execute(ctx, g.ID(), hostID, currentTurnID)
	testutil.AssertNoError(t, err, "Should kick current turn player")

	newTurnID := g.CurrentTurn().PlayerID()
	testutil.AssertNotEqual(t, currentTurnID, newTurnID, "Turn should advance to different player")

	newTurnPlayer, _ := g.GetPlayer(newTurnID)
	testutil.AssertFalse(t, newTurnPlayer.HasExited(), "New turn player should not be exited")
}

func TestKickPlayer_TheirTurn_AllPassedTriggersProduction(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()

	finalScoring := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, testutil.TestLogger())
	kick := newKickAction(repo, finalScoring)

	hostID := g.HostPlayerID()

	// Find non-host player and make it their turn
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}
	_ = g.SetCurrentTurn(ctx, targetID, 2)

	// Host has already passed
	hostPlayer, _ := g.GetPlayer(hostID)
	hostPlayer.SetPassed(true)

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick last non-passed player")

	// Should have triggered production phase (or end of generation)
	phase := g.CurrentPhase()
	testutil.AssertTrue(t, phase == shared.GamePhaseProductionAndCardDraw || phase == shared.GamePhaseComplete,
		"Should transition to production or complete phase, got: "+string(phase))
}

func TestKickPlayer_StartingSelection_AdvancesGame(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	// Set game to starting selection phase
	_ = g.UpdateStatus(ctx, shared.GameStatusActive)
	_ = g.UpdatePhase(ctx, shared.GamePhaseStartingSelection)

	hostID := g.HostPlayerID()
	players := g.GetAllPlayers()
	var targetID string
	for _, p := range players {
		if p.ID() != hostID {
			targetID = p.ID()
			break
		}
	}

	// Set turn order
	_ = g.SetTurnOrder(ctx, []string{hostID, targetID})

	// Both players have pending starting selections
	_ = g.SetSelectCorporationPhase(ctx, hostID, &shared.SelectCorporationPhase{
		AvailableCorporations: []string{"corp1"},
	})
	_ = g.SetSelectCorporationPhase(ctx, targetID, &shared.SelectCorporationPhase{
		AvailableCorporations: []string{"corp2"},
	})
	_ = g.SetSelectStartingCardsPhase(ctx, hostID, &shared.SelectStartingCardsPhase{
		AvailableCards: []string{"card1"},
	})
	_ = g.SetSelectStartingCardsPhase(ctx, targetID, &shared.SelectStartingCardsPhase{
		AvailableCards: []string{"card2"},
	})

	// Clear host's phases (host has completed selection)
	_ = g.SetSelectCorporationPhase(ctx, hostID, nil)
	_ = g.SetSelectStartingCardsPhase(ctx, hostID, nil)

	// Target still has pending phases - kick them
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player during starting selection")

	// Since all remaining (non-exited) players are done, should advance to action phase
	testutil.AssertEqual(t, shared.GamePhaseAction, g.CurrentPhase(), "Should advance to action phase")
}

func TestKickPlayer_StartingSelection_WaitsForOthers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	_ = g.UpdateStatus(ctx, shared.GameStatusActive)
	_ = g.UpdatePhase(ctx, shared.GamePhaseStartingSelection)

	hostID := g.HostPlayerID()
	players := g.GetAllPlayers()
	var targetID, thirdPlayerID string
	for _, p := range players {
		if p.ID() != hostID {
			if targetID == "" {
				targetID = p.ID()
			} else {
				thirdPlayerID = p.ID()
			}
		}
	}
	_ = g.SetTurnOrder(ctx, []string{hostID, targetID, thirdPlayerID})

	// All three have pending selection
	for _, id := range []string{hostID, targetID, thirdPlayerID} {
		_ = g.SetSelectCorporationPhase(ctx, id, &shared.SelectCorporationPhase{
			AvailableCorporations: []string{"corp1"},
		})
	}

	// Kick target - but third player still has pending selection
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Should NOT advance because third player still has pending selection
	testutil.AssertEqual(t, shared.GamePhaseStartingSelection, g.CurrentPhase(), "Should remain in starting selection")
}

func TestKickPlayer_ProductionPhase_AdvancesWhenAllDone(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	// Set phase to production
	_ = g.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Host has completed production, target has not
	_ = g.SetProductionPhase(ctx, hostID, &shared.ProductionPhase{
		SelectionComplete: true,
	})
	_ = g.SetProductionPhase(ctx, targetID, &shared.ProductionPhase{
		SelectionComplete: false,
		AvailableCards:    []string{"card1"},
	})

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player in production phase")

	// Since host (only remaining) has completed, should advance to action phase
	testutil.AssertEqual(t, shared.GamePhaseAction, g.CurrentPhase(), "Should advance to action phase")
}

func TestKickPlayer_PendingTileSelection_Cleared(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Give target a pending tile selection
	_ = g.SetPendingTileSelection(ctx, targetID, &shared.PendingTileSelection{
		TileType:       "city",
		AvailableHexes: []string{"hex1", "hex2"},
		Source:         "test",
	})

	// Set current turn to host so kicking target doesn't need turn advancement
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player with pending tile selection")

	// Tile selection should be cleared
	tileSelection := g.GetPendingTileSelection(targetID)
	testutil.AssertTrue(t, tileSelection == nil, "Pending tile selection should be cleared")
}

func TestKickPlayer_ExitedPlayerCannotReconnect(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Set current turn to host
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Try to reconnect via takeover
	takeoverAction := connection.NewPlayerTakeoverAction(repo, cardRegistry, testutil.TestLogger())
	_, err = takeoverAction.Execute(ctx, g.ID(), targetID)
	testutil.AssertError(t, err, "Exited player should not be able to reconnect")
}

func TestKickPlayer_TurnSkipsExitedPlayer(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()

	// Ensure turn order is predictable: [hostID, targetID, thirdID]
	var targetID, thirdID string
	for _, id := range playerIDs {
		if id != hostID {
			if targetID == "" {
				targetID = id
			} else {
				thirdID = id
			}
		}
	}
	_ = g.SetTurnOrder(ctx, []string{hostID, targetID, thirdID})
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	// Kick the middle player
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Verify the exited player is in the turn order but at the end
	turnOrder := g.TurnOrder()
	testutil.AssertEqual(t, targetID, turnOrder[len(turnOrder)-1], "Exited player should be at end")

	// The exited player should have HasPassed() true, so skip logic will skip them
	exitedPlayer, _ := g.GetPlayer(targetID)
	testutil.AssertTrue(t, exitedPlayer.HasPassed(), "Exited player should be passed")
	testutil.AssertTrue(t, exitedPlayer.HasExited(), "Exited player should be exited")
}

func TestKickPlayer_ProductionSkipsExitedPlayer(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Set current turn to host
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	// Kick the target
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Exited player should not have a production phase set
	productionPhase := g.GetProductionPhase(targetID)
	testutil.AssertTrue(t, productionPhase == nil, "Exited player should not have production phase data")
}

// ============================================================================
// Validation
// ============================================================================

func TestKickPlayer_OnlyHostCanKick(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var nonHostRequester, nonHostTarget string
	for _, id := range playerIDs {
		if id != hostID {
			if nonHostRequester == "" {
				nonHostRequester = id
			} else {
				nonHostTarget = id
			}
		}
	}

	err := kick.Execute(ctx, g.ID(), nonHostRequester, nonHostTarget)
	testutil.AssertError(t, err, "Non-host should not be able to kick")

	target, _ := g.GetPlayer(nonHostTarget)
	testutil.AssertFalse(t, target.HasExited(), "Player should not be exited")
}

func TestKickPlayer_CannotKickSelf(t *testing.T) {
	g, repo, _, _ := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()

	err := kick.Execute(ctx, g.ID(), hostID, hostID)
	testutil.AssertError(t, err, "Should not be able to kick yourself")
}

func TestKickPlayer_CannotKickAlreadyExited(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Set current turn to host
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	// Kick once
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "First kick should succeed")

	// Kick again - should fail
	err = kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertError(t, err, "Should not be able to kick already exited player")
}

func TestKickPlayer_LobbyKickRemovesPlayer(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	players := g.GetAllPlayers()
	var targetID string
	for _, p := range players {
		if p.ID() != hostID {
			targetID = p.ID()
			break
		}
	}

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Lobby kick should succeed")

	remaining := g.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(remaining), "Player should be removed entirely from lobby")

	for _, p := range remaining {
		testutil.AssertNotEqual(t, targetID, p.ID(), "Kicked player should not be in game")
	}
}

func TestKickPlayer_ExitedPlayerIncludedInScoring(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Set current turn to host
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Exited player should still exist in the game
	exitedPlayer, err := g.GetPlayer(targetID)
	testutil.AssertNoError(t, err, "Exited player should still be gettable")
	testutil.AssertTrue(t, exitedPlayer.HasExited(), "Player should be marked as exited")

	// All players (including exited) should be in GetAllPlayers for scoring
	allPlayers := g.GetAllPlayers()
	found := false
	for _, p := range allPlayers {
		if p.ID() == targetID {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Exited player should be included in all players for scoring")
}

func TestKickPlayer_ExitedPlayerStaysAtBottomAcrossGenerations(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()

	finalScoring := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, testutil.TestLogger())
	kick := newKickAction(repo, finalScoring)

	hostID := g.HostPlayerID()
	var targetID, thirdID string
	for _, id := range playerIDs {
		if id != hostID {
			if targetID == "" {
				targetID = id
			} else {
				thirdID = id
			}
		}
	}

	// Set turn order: [hostID, targetID, thirdID]
	_ = g.SetTurnOrder(ctx, []string{hostID, targetID, thirdID})
	_ = g.SetCurrentTurn(ctx, hostID, 2)

	// Kick the middle player
	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player")

	// Turn order should now be [hostID, thirdID, targetID(exited)]
	turnOrder := g.TurnOrder()
	testutil.AssertEqual(t, targetID, turnOrder[len(turnOrder)-1], "Exited player should be at end after kick")

	// Simulate a generation ending: both remaining players pass → production
	hostPlayer, _ := g.GetPlayer(hostID)
	thirdPlayer, _ := g.GetPlayer(thirdID)
	hostPlayer.SetPassed(true)
	thirdPlayer.SetPassed(true)

	activePlayers := []*player.Player{hostPlayer, thirdPlayer}
	err = turn_management.ExecuteProductionPhase(ctx, g, activePlayers, testutil.TestLogger())
	testutil.AssertNoError(t, err, "Production phase should succeed")

	// After production phase rotates turn order, exited player should STILL be at the end
	turnOrder = g.TurnOrder()
	testutil.AssertEqual(t, targetID, turnOrder[len(turnOrder)-1],
		"Exited player should remain at end of turn order after generation advance")

	// Active players should have rotated among themselves
	// Before rotation: [hostID, thirdID, targetID(exited)]
	// After rotation: [thirdID, hostID, targetID(exited)]
	testutil.AssertEqual(t, thirdID, turnOrder[0], "First active player should rotate to front")
	testutil.AssertEqual(t, hostID, turnOrder[1], "Second active player should be previous first")
}

func TestKickPlayer_TheirTurn_LastActiveGetsUnlimitedActions(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	kick := newKickAction(repo, nil)

	hostID := g.HostPlayerID()
	var targetID, thirdID string
	for _, id := range playerIDs {
		if id != hostID {
			if targetID == "" {
				targetID = id
			} else {
				thirdID = id
			}
		}
	}

	// Set turn order: [hostID, targetID, thirdID]
	_ = g.SetTurnOrder(ctx, []string{hostID, targetID, thirdID})

	// Host has passed, third player is active
	hostPlayer, _ := g.GetPlayer(hostID)
	hostPlayer.SetPassed(true)

	// It's the target's turn
	_ = g.SetCurrentTurn(ctx, targetID, 2)

	err := kick.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Should kick player whose turn it is")

	// Third player should now have the turn with unlimited actions
	currentTurn := g.CurrentTurn()
	testutil.AssertEqual(t, thirdID, currentTurn.PlayerID(), "Third player should have the turn")
	testutil.AssertEqual(t, -1, currentTurn.ActionsRemaining(), "Last active player should get unlimited actions")
}
