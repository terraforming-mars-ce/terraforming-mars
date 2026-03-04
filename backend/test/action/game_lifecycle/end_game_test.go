package game_lifecycle_test

import (
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

func newEndGameAction(repo game.GameRepository) *connection.EndGameAction {
	return connection.NewEndGameAction(repo, nil, testutil.TestLogger())
}

// ============================================================================
// Green Path
// ============================================================================

func TestEndGame_HostCanEndLobbyGame(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()
	action := newEndGameAction(repo)

	hostID := g.HostPlayerID()
	gameID := g.ID()

	err := action.Execute(ctx, gameID, hostID)
	testutil.AssertNoError(t, err, "Host should be able to end game")

	testutil.AssertFalse(t, repo.Exists(ctx, gameID), "Game should be deleted from repository")
}

func TestEndGame_HostCanEndActiveGame(t *testing.T) {
	g, repo, _, _ := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	action := newEndGameAction(repo)

	hostID := g.HostPlayerID()
	gameID := g.ID()

	err := action.Execute(ctx, gameID, hostID)
	testutil.AssertNoError(t, err, "Host should be able to end active game")

	testutil.AssertFalse(t, repo.Exists(ctx, gameID), "Game should be deleted")
}

// ============================================================================
// Validation
// ============================================================================

func TestEndGame_NonHostCannotEndGame(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	action := newEndGameAction(repo)

	hostID := g.HostPlayerID()
	var nonHostID string
	for _, id := range playerIDs {
		if id != hostID {
			nonHostID = id
			break
		}
	}

	err := action.Execute(ctx, g.ID(), nonHostID)
	testutil.AssertError(t, err, "Non-host should not be able to end game")

	testutil.AssertTrue(t, repo.Exists(ctx, g.ID()), "Game should still exist")
}

func TestEndGame_InvalidGameID(t *testing.T) {
	_, repo, _, _ := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	action := newEndGameAction(repo)

	err := action.Execute(ctx, "nonexistent-game", "some-player")
	testutil.AssertError(t, err, "Should fail for nonexistent game")
}
