package game_lifecycle_test

import (
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func newSetPlayerColorAction(repo game.GameRepository) *connection.SetPlayerColorAction {
	return connection.NewSetPlayerColorAction(repo, testutil.TestLogger())
}

// ============================================================================
// Auto-assign on Join
// ============================================================================

func TestPlayerColor_AutoAssignedOnJoin(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)

	players := g.GetAllPlayers()
	colors := make(map[string]bool)
	for _, p := range players {
		testutil.AssertTrue(t, p.Color() != "", "Player should have an auto-assigned color")
		colors[p.Color()] = true
	}
	testutil.AssertEqual(t, 3, len(colors), "Each player should have a unique color")
}

func TestPlayerColor_AssignedFromPalette(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)

	paletteSet := make(map[string]bool)
	for _, c := range shared.PlayerColors {
		paletteSet[c] = true
	}

	for _, p := range g.GetAllPlayers() {
		testutil.AssertTrue(t, paletteSet[p.Color()], "Player color should be from the palette")
	}
}

// ============================================================================
// SetPlayerColor - Green Path
// ============================================================================

func TestSetPlayerColor_Success(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	players := g.GetAllPlayers()
	player := players[0]
	otherPlayer := players[1]

	var availableColor string
	for _, c := range shared.PlayerColors {
		if c != player.Color() && c != otherPlayer.Color() {
			availableColor = c
			break
		}
	}

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player.ID(), player.ID(), availableColor)

	testutil.AssertNoError(t, err, "Should set player color")
	testutil.AssertEqual(t, availableColor, player.Color(), "Player color should be updated")
}

func TestSetPlayerColor_OwnColorIsNoop(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	player := g.GetAllPlayers()[0]
	originalColor := player.Color()

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player.ID(), player.ID(), originalColor)

	testutil.AssertNoError(t, err, "Re-selecting own color should succeed")
	testutil.AssertEqual(t, originalColor, player.Color(), "Color should remain the same")
}

func TestSetPlayerColor_HostCanChangeBotColor(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := testutil.TestContext()

	hostPlayer := g.GetAllPlayers()[0]
	bot := testutil.AddBotToGame(t, g, repo, testutil.CreateTestCardRegistry(), "TestBot", "normal", "fast")

	var availableColor string
	for _, c := range shared.PlayerColors {
		if c != hostPlayer.Color() && c != bot.Color() {
			availableColor = c
			break
		}
	}

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), hostPlayer.ID(), bot.ID(), availableColor)

	testutil.AssertNoError(t, err, "Host should change bot color")
	testutil.AssertEqual(t, availableColor, bot.Color(), "Bot color should be updated")
}

// ============================================================================
// SetPlayerColor - Red Path
// ============================================================================

func TestSetPlayerColor_NotInLobby(t *testing.T) {
	g, repo, _, player1ID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := testutil.TestContext()

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player1ID, player1ID, shared.PlayerColors[0])

	testutil.AssertError(t, err, "Should reject color change when not in lobby")
}

func TestSetPlayerColor_ColorTaken(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	players := g.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player1.ID(), player1.ID(), player2.Color())

	testutil.AssertError(t, err, "Should reject color that is taken by another player")
}

func TestSetPlayerColor_InvalidColor(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	player := g.GetAllPlayers()[0]

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player.ID(), player.ID(), "#123456")

	testutil.AssertError(t, err, "Should reject color not in the palette")
}

func TestSetPlayerColor_NonHostCannotChangeBotColor(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	nonHost, _ := g.GetPlayer("player-2")
	bot := testutil.AddBotToGame(t, g, repo, testutil.CreateTestCardRegistry(), "TestBot", "normal", "fast")

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), nonHost.ID(), bot.ID(), shared.PlayerColors[5])

	testutil.AssertError(t, err, "Non-host should not change bot color")
}

func TestSetPlayerColor_CannotChangeOtherHumanColor(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	host, _ := g.GetPlayer("player-1")
	otherHuman, _ := g.GetPlayer("player-2")

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), host.ID(), otherHuman.ID(), shared.PlayerColors[5])

	testutil.AssertError(t, err, "Host should not change other human player's color")
}

func TestSetPlayerColor_GameNotFound(t *testing.T) {
	repo := testutil.NewTestGameRepository(t)
	action := newSetPlayerColorAction(repo)

	err := action.Execute(testutil.TestContext(), "nonexistent", "player-1", "player-1", shared.PlayerColors[0])
	testutil.AssertError(t, err, "Should fail for nonexistent game")
}

func TestSetPlayerColor_PlayerNotFound(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	player := g.GetAllPlayers()[0]

	action := newSetPlayerColorAction(repo)
	err := action.Execute(ctx, g.ID(), player.ID(), "nonexistent", shared.PlayerColors[0])

	testutil.AssertError(t, err, "Should fail for nonexistent player")
}

// ============================================================================
// Colors Preserved After Game Start
// ============================================================================

func TestPlayerColor_PreservedAfterStart(t *testing.T) {
	g, repo, _, player1ID, player2ID := testutil.SetupTwoPlayerGame(t)

	p1, _ := g.GetPlayer(player1ID)
	p2, _ := g.GetPlayer(player2ID)

	testutil.AssertTrue(t, p1.Color() != "", "Player 1 should have color after start")
	testutil.AssertTrue(t, p2.Color() != "", "Player 2 should have color after start")
	testutil.AssertNotEqual(t, p1.Color(), p2.Color(), "Players should have different colors")
	_ = repo
}
