package game_lifecycle_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

// ============================================================================
// KickPlayerAction Tests
// ============================================================================

func TestKickPlayerAction_HostKicksNonHostInLobby_Success(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := kickAction.Execute(context.Background(), testGame.ID(), hostPlayerID, nonHostPlayerID)

	testutil.AssertNoError(t, err, "Host should be able to kick non-host player")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(remainingPlayers), "Should have 2 players remaining")

	for _, p := range remainingPlayers {
		testutil.AssertNotEqual(t, nonHostPlayerID, p.ID(), "Kicked player should not be in game")
	}
}

func TestKickPlayerAction_NonHostCannotKick_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostRequester, nonHostTarget string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			if nonHostRequester == "" {
				nonHostRequester = p.ID()
			} else if nonHostTarget == "" {
				nonHostTarget = p.ID()
				break
			}
		}
	}

	err := kickAction.Execute(context.Background(), testGame.ID(), nonHostRequester, nonHostTarget)

	testutil.AssertError(t, err, "Non-host should not be able to kick players")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 3, len(fetchedGame.GetAllPlayers()), "No player should be removed")
}

func TestKickPlayerAction_HostCannotKickSelf_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()

	err := kickAction.Execute(context.Background(), testGame.ID(), hostPlayerID, hostPlayerID)

	testutil.AssertError(t, err, "Host should not be able to kick themselves")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No player should be removed")
}

func TestKickPlayerAction_CannotKickInActiveGame_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	testutil.StartTestGame(t, testGame)

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := kickAction.Execute(context.Background(), testGame.ID(), hostPlayerID, nonHostPlayerID)

	testutil.AssertError(t, err, "Should not be able to kick players in active game")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No player should be removed")
}

func TestKickPlayerAction_GameNotFound_Error(t *testing.T) {
	repo := game.NewInMemoryGameRepository()
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	err := kickAction.Execute(context.Background(), "non-existent-game", "host-id", "target-id")

	testutil.AssertError(t, err, "Should fail when game doesn't exist")
}

func TestKickPlayerAction_PlayerNotFound_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()

	err := kickAction.Execute(context.Background(), testGame.ID(), hostPlayerID, "non-existent-player")

	testutil.AssertError(t, err, "Should fail when target player doesn't exist")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No player should be removed")
}

func TestKickPlayerAction_KickLeavesHostAlone_Success(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	kickAction := connection.NewKickPlayerAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := kickAction.Execute(context.Background(), testGame.ID(), hostPlayerID, nonHostPlayerID)

	testutil.AssertNoError(t, err, "Host should be able to kick the last non-host player")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(remainingPlayers), "Should have only host remaining")
	testutil.AssertEqual(t, hostPlayerID, remainingPlayers[0].ID(), "Host should remain")
	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should still be host")
}

// ============================================================================
// PlayerDisconnectedAction Tests
// ============================================================================

func TestPlayerDisconnectedAction_LobbyRegularPlayerLeaves_Removed(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := disconnectAction.Execute(context.Background(), testGame.ID(), nonHostPlayerID)

	testutil.AssertNoError(t, err, "Non-host leaving lobby should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(remainingPlayers), "Should have 2 players remaining")

	for _, p := range remainingPlayers {
		testutil.AssertNotEqual(t, nonHostPlayerID, p.ID(), "Disconnected player should be removed")
	}

	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should remain unchanged")
}

func TestPlayerDisconnectedAction_LobbyHostLeaves_HostReassigned(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()

	err := disconnectAction.Execute(context.Background(), testGame.ID(), hostPlayerID)

	testutil.AssertNoError(t, err, "Host leaving lobby should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(remainingPlayers), "Should have 2 players remaining")

	testutil.AssertNotEqual(t, "", fetchedGame.HostPlayerID(), "New host should be assigned")
	testutil.AssertNotEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "New host should be different from old host")

	var foundNewHost bool
	for _, p := range remainingPlayers {
		if p.ID() == fetchedGame.HostPlayerID() {
			foundNewHost = true
			break
		}
	}
	testutil.AssertTrue(t, foundNewHost, "New host should be one of the remaining players")
}

func TestPlayerDisconnectedAction_LobbyHostLeavesAsLastPlayer_GameDeleted(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	gameID := testGame.ID()

	err := disconnectAction.Execute(context.Background(), gameID, hostPlayerID)

	testutil.AssertNoError(t, err, "Last player leaving should succeed")

	_, err = repo.Get(context.Background(), gameID)
	testutil.AssertError(t, err, "Game should be deleted after last player leaves")
}

func TestPlayerDisconnectedAction_LobbyLastNonHostLeaves_HostAlone(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := disconnectAction.Execute(context.Background(), testGame.ID(), nonHostPlayerID)

	testutil.AssertNoError(t, err, "Last non-host leaving should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(remainingPlayers), "Should have only host remaining")
	testutil.AssertEqual(t, hostPlayerID, remainingPlayers[0].ID(), "Host should remain")
	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should still be host")
}

func TestPlayerDisconnectedAction_ActiveGamePlayerDisconnects_MarkedDisconnected(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	testutil.StartTestGame(t, testGame)

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostPlayerID string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostPlayerID = p.ID()
			break
		}
	}

	err := disconnectAction.Execute(context.Background(), testGame.ID(), nonHostPlayerID)

	testutil.AssertNoError(t, err, "Player disconnecting from active game should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	remainingPlayers := fetchedGame.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(remainingPlayers), "Player should not be removed in active game")

	player, _ := fetchedGame.GetPlayer(nonHostPlayerID)
	testutil.AssertFalse(t, player.IsConnected(), "Player should be marked as disconnected")
}

func TestPlayerDisconnectedAction_ActiveGameHostDisconnects_NoReassignment(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	testutil.StartTestGame(t, testGame)

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()

	err := disconnectAction.Execute(context.Background(), testGame.ID(), hostPlayerID)

	testutil.AssertNoError(t, err, "Host disconnecting from active game should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should not change in active game")

	player, _ := fetchedGame.GetPlayer(hostPlayerID)
	testutil.AssertFalse(t, player.IsConnected(), "Host should be marked as disconnected")

	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No players should be removed")
}

func TestPlayerDisconnectedAction_GameNotFound_Error(t *testing.T) {
	repo := game.NewInMemoryGameRepository()
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	err := disconnectAction.Execute(context.Background(), "non-existent-game", "player-id")

	testutil.AssertError(t, err, "Should fail when game doesn't exist")
}

func TestPlayerDisconnectedAction_PlayerNotFoundInLobby_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	err := disconnectAction.Execute(context.Background(), testGame.ID(), "non-existent-player")

	testutil.AssertError(t, err, "Should fail when player doesn't exist in lobby")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No player should be removed")
}

func TestPlayerDisconnectedAction_PlayerNotFoundInActiveGame_Error(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()

	testutil.StartTestGame(t, testGame)

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	err := disconnectAction.Execute(context.Background(), testGame.ID(), "non-existent-player")

	testutil.AssertError(t, err, "Should fail when player doesn't exist in active game")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "No player should be affected")
}

func TestPlayerDisconnectedAction_MultiplePlayersLeavingSequence_Lobby(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 4, broadcaster)
	logger := testutil.TestLogger()

	disconnectAction := connection.NewPlayerDisconnectedAction(repo, logger)

	hostPlayerID := testGame.HostPlayerID()
	players := testGame.GetAllPlayers()
	var nonHostIDs []string
	for _, p := range players {
		if p.ID() != hostPlayerID {
			nonHostIDs = append(nonHostIDs, p.ID())
		}
	}

	err := disconnectAction.Execute(context.Background(), testGame.ID(), nonHostIDs[0])
	testutil.AssertNoError(t, err, "First non-host leaving should succeed")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 3, len(fetchedGame.GetAllPlayers()), "Should have 3 players")
	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should remain")

	err = disconnectAction.Execute(context.Background(), testGame.ID(), nonHostIDs[1])
	testutil.AssertNoError(t, err, "Second non-host leaving should succeed")

	fetchedGame, _ = repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 2, len(fetchedGame.GetAllPlayers()), "Should have 2 players")
	testutil.AssertEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "Host should remain")

	err = disconnectAction.Execute(context.Background(), testGame.ID(), hostPlayerID)
	testutil.AssertNoError(t, err, "Host leaving should succeed with reassignment")

	fetchedGame, _ = repo.Get(context.Background(), testGame.ID())
	testutil.AssertEqual(t, 1, len(fetchedGame.GetAllPlayers()), "Should have 1 player")
	testutil.AssertNotEqual(t, hostPlayerID, fetchedGame.HostPlayerID(), "New host should be assigned")
	testutil.AssertEqual(t, nonHostIDs[2], fetchedGame.HostPlayerID(), "Remaining player should be host")

	err = disconnectAction.Execute(context.Background(), testGame.ID(), nonHostIDs[2])
	testutil.AssertNoError(t, err, "Last player leaving should succeed")

	_, err = repo.Get(context.Background(), testGame.ID())
	testutil.AssertError(t, err, "Game should be deleted")
}
