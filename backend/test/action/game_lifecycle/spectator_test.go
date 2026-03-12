package game_lifecycle_test

import (
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func newSpectateAction(repo game.GameRepository) *connection.SpectateGameAction {
	return connection.NewSpectateGameAction(repo, testutil.TestLogger())
}

func newSpectatorDisconnectedAction(repo game.GameRepository) *connection.SpectatorDisconnectedAction {
	return connection.NewSpectatorDisconnectedAction(repo, testutil.TestLogger())
}

func newKickSpectatorAction(repo game.GameRepository) *connection.KickSpectatorAction {
	return connection.NewKickSpectatorAction(repo, testutil.TestLogger())
}

// ============================================================================
// SpectateGame - Green Path
// ============================================================================

func TestSpectateGame_LobbyGame(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSpectateAction(repo)
	result, err := action.Execute(ctx, g.ID(), "Spectator1", "spec-1")

	testutil.AssertNoError(t, err, "Should join lobby game as spectator")
	testutil.AssertEqual(t, "spec-1", result.SpectatorID, "Spectator ID mismatch")
	testutil.AssertEqual(t, 1, g.SpectatorCount(), "Should have 1 spectator")
}

func TestSpectateGame_ActiveGame(t *testing.T) {
	g, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	ctx := testutil.TestContext()

	action := newSpectateAction(repo)
	result, err := action.Execute(ctx, g.ID(), "Watcher", "spec-2")

	testutil.AssertNoError(t, err, "Should join active game as spectator")
	testutil.AssertEqual(t, "spec-2", result.SpectatorID, "Spectator ID mismatch")
	testutil.AssertEqual(t, 1, g.SpectatorCount(), "Should have 1 spectator")
}

func TestSpectateGame_MultipleSpectators(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSpectateAction(repo)
	for i := 0; i < 4; i++ {
		_, err := action.Execute(ctx, g.ID(), "Spec", "spec-"+string(rune('a'+i)))
		testutil.AssertNoError(t, err, "Should join as spectator")
	}

	testutil.AssertEqual(t, 4, g.SpectatorCount(), "Should have 4 spectators")
}

func TestSpectateGame_ColorAssignment(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSpectateAction(repo)
	_, _ = action.Execute(ctx, g.ID(), "Spec1", "spec-1")
	_, _ = action.Execute(ctx, g.ID(), "Spec2", "spec-2")

	specs := g.GetAllSpectators()
	testutil.AssertEqual(t, 2, len(specs), "Should have 2 spectators")

	colors := make(map[string]bool)
	for _, s := range specs {
		testutil.AssertTrue(t, s.Color() != "", "Spectator should have a color")
		colors[s.Color()] = true
	}
	testutil.AssertEqual(t, 2, len(colors), "Each spectator should have a unique color")
}

// ============================================================================
// SpectateGame - Red Path
// ============================================================================

func TestSpectateGame_MaxSpectators(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSpectateAction(repo)
	for i := 0; i < shared.MaxSpectators; i++ {
		_, err := action.Execute(ctx, g.ID(), "Spec", "spec-"+string(rune('a'+i)))
		testutil.AssertNoError(t, err, "Should join as spectator")
	}

	_, err := action.Execute(ctx, g.ID(), "TooMany", "spec-extra")
	testutil.AssertError(t, err, "Should reject 5th spectator")
}

func TestSpectateGame_GameNotFound(t *testing.T) {
	repo := testutil.NewTestGameRepository(t)
	action := newSpectateAction(repo)

	_, err := action.Execute(testutil.TestContext(), "nonexistent", "Spec", "spec-1")
	testutil.AssertError(t, err, "Should fail for nonexistent game")
}

// ============================================================================
// SpectatorDisconnected - Green Path
// ============================================================================

func TestSpectatorDisconnected_RemovedFromGame(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec1", "spec-1")
	testutil.AssertEqual(t, 1, g.SpectatorCount(), "Should have 1 spectator")

	disconnectAction := newSpectatorDisconnectedAction(repo)
	err := disconnectAction.Execute(ctx, g.ID(), "spec-1")
	testutil.AssertNoError(t, err, "Should disconnect spectator")
	testutil.AssertEqual(t, 0, g.SpectatorCount(), "Spectator should be fully removed")
}

func TestSpectatorDisconnected_DoesNotAffectPlayers(t *testing.T) {
	g, repo, _, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec", "spec-1")

	disconnectAction := newSpectatorDisconnectedAction(repo)
	_ = disconnectAction.Execute(ctx, g.ID(), "spec-1")

	players := g.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(players), "Players should be unaffected")
	_, err1 := g.GetPlayer(player1ID)
	_, err2 := g.GetPlayer(player2ID)
	testutil.AssertNoError(t, err1, "Player 1 should still exist")
	testutil.AssertNoError(t, err2, "Player 2 should still exist")
}

// ============================================================================
// KickSpectator - Green Path
// ============================================================================

func TestKickSpectator_HostCanKick(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec1", "spec-1")

	kickAction := newKickSpectatorAction(repo)
	err := kickAction.Execute(ctx, g.ID(), g.HostPlayerID(), "spec-1")
	testutil.AssertNoError(t, err, "Host should be able to kick spectator")
	testutil.AssertEqual(t, 0, g.SpectatorCount(), "Spectator should be removed")
}

// ============================================================================
// KickSpectator - Red Path
// ============================================================================

func TestKickSpectator_NonHostCannotKick(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec1", "spec-1")

	players := g.GetAllPlayers()
	var nonHostID string
	for _, p := range players {
		if p.ID() != g.HostPlayerID() {
			nonHostID = p.ID()
			break
		}
	}

	kickAction := newKickSpectatorAction(repo)
	err := kickAction.Execute(ctx, g.ID(), nonHostID, "spec-1")
	testutil.AssertError(t, err, "Non-host should not be able to kick spectator")
	testutil.AssertEqual(t, 1, g.SpectatorCount(), "Spectator should remain")
}

func TestKickSpectator_SpectatorNotFound(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	kickAction := newKickSpectatorAction(repo)
	err := kickAction.Execute(ctx, g.ID(), g.HostPlayerID(), "nonexistent")
	testutil.AssertError(t, err, "Should fail for nonexistent spectator")
}

// ============================================================================
// Spectator Not In Players
// ============================================================================

func TestSpectator_NotInGetAllPlayers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec1", "spec-1")

	players := g.GetAllPlayers()
	for _, p := range players {
		testutil.AssertNotEqual(t, "spec-1", p.ID(), "Spectator should not appear in players")
	}
}
