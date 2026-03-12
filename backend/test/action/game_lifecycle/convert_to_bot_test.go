package game_lifecycle_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func newConvertAction(repo game.GameRepository) *gameaction.ConvertToBotAction {
	return gameaction.NewConvertToBotAction(repo, nil, testutil.TestLogger())
}

// ============================================================================
// Green Path
// ============================================================================

func TestConvertToBot_Success(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	convert := newConvertAction(repo)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	err := convert.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "Convert to bot should succeed")

	target, _ := g.GetPlayer(targetID)
	testutil.AssertTrue(t, target.IsBot(), "Player should be a bot")
	testutil.AssertEqual(t, player.BotStatusLoading, target.BotStatus(), "Bot status should be loading")
	testutil.AssertEqual(t, player.BotDifficultyNormal, target.BotDifficulty(), "Bot difficulty should be normal")
	testutil.AssertEqual(t, player.BotSpeedFast, target.BotSpeed(), "Bot speed should be fast")
	testutil.AssertTrue(t, target.IsConnected(), "Bot should be connected")
}

// ============================================================================
// Validation
// ============================================================================

func TestConvertToBot_NotHost_Fails(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	convert := newConvertAction(repo)

	hostID := g.HostPlayerID()
	var nonHostRequester, target string
	for _, id := range playerIDs {
		if id != hostID {
			if nonHostRequester == "" {
				nonHostRequester = id
			} else {
				target = id
			}
		}
	}

	err := convert.Execute(ctx, g.ID(), nonHostRequester, target)
	testutil.AssertError(t, err, "Non-host should not be able to convert")
}

func TestConvertToBot_Self_Fails(t *testing.T) {
	g, repo, _, _ := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	convert := newConvertAction(repo)
	hostID := g.HostPlayerID()

	err := convert.Execute(ctx, g.ID(), hostID, hostID)
	testutil.AssertError(t, err, "Should not be able to convert yourself")
}

func TestConvertToBot_AlreadyBot_Fails(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	convert := newConvertAction(repo)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Convert once
	err := convert.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertNoError(t, err, "First convert should succeed")

	// Convert again - should fail
	err = convert.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertError(t, err, "Should not convert already-bot player")
}

func TestConvertToBot_Exited_Fails(t *testing.T) {
	g, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	convert := newConvertAction(repo)

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Mark as exited
	target, _ := g.GetPlayer(targetID)
	target.SetExited(true)

	err := convert.Execute(ctx, g.ID(), hostID, targetID)
	testutil.AssertError(t, err, "Should not convert exited player")
}

// ============================================================================
// Takeover Blocked for Bots
// ============================================================================

func TestPlayerTakeover_BlockedForBot(t *testing.T) {
	g, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	ctx := testutil.TestContext()
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	hostID := g.HostPlayerID()
	var targetID string
	for _, id := range playerIDs {
		if id != hostID {
			targetID = id
			break
		}
	}

	// Convert to bot
	target, _ := g.GetPlayer(targetID)
	target.SetPlayerType(player.PlayerTypeBot)
	target.SetConnected(false)

	// Try takeover
	takeoverAction := connection.NewPlayerTakeoverAction(repo, cardRegistry, testutil.TestLogger())
	_, err := takeoverAction.Execute(ctx, g.ID(), targetID)
	testutil.AssertError(t, err, "Should not be able to take over a bot player")
}

// ============================================================================
// Bot-Aware Host Reassignment
// ============================================================================

func TestGameDeletedWhenLastHumanLeavesLobbyWithBots(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	// Add a bot player manually
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})
	cardRegistry := testutil.CreateTestCardRegistry()
	addBotAction := gameaction.NewAddBotAction(repo, cardRegistry, nil, nil, testutil.TestLogger())
	_, err := addBotAction.Execute(ctx, g.ID(), "", "", "")
	testutil.AssertNoError(t, err, "Should add bot")

	// Verify we have 1 human + 1 bot
	players := g.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(players), "Should have 2 players")

	hostID := g.HostPlayerID()

	// Host (last human) disconnects
	disconnectAction := connection.NewPlayerDisconnectedAction(repo, testutil.TestLogger())
	err = disconnectAction.Execute(ctx, g.ID(), hostID)
	testutil.AssertNoError(t, err, "Disconnect should succeed")

	// Game should be deleted since only bots remain
	_, err = repo.Get(ctx, g.ID())
	testutil.AssertError(t, err, "Game should be deleted when last human leaves")
}

func TestHostReassignedToHumanNotBot(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := context.Background()

	// Add a bot player
	g.UpdateSettings(ctx, shared.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})
	cardRegistry := testutil.CreateTestCardRegistry()
	addBotAction := gameaction.NewAddBotAction(repo, cardRegistry, nil, nil, testutil.TestLogger())
	_, err := addBotAction.Execute(ctx, g.ID(), "", "", "")
	testutil.AssertNoError(t, err, "Should add bot")

	// We have: host (human), player2 (human), bot
	hostID := g.HostPlayerID()
	var humanID string
	for _, p := range g.GetAllPlayers() {
		if p.ID() != hostID && !p.IsBot() {
			humanID = p.ID()
			break
		}
	}
	testutil.AssertTrue(t, humanID != "", "Should find second human player")

	// Host disconnects
	disconnectAction := connection.NewPlayerDisconnectedAction(repo, testutil.TestLogger())
	err = disconnectAction.Execute(ctx, g.ID(), hostID)
	testutil.AssertNoError(t, err, "Disconnect should succeed")

	// Game should still exist
	updatedGame, err := repo.Get(ctx, g.ID())
	testutil.AssertNoError(t, err, "Game should still exist")

	// New host should be the human, not the bot
	newHostID := updatedGame.HostPlayerID()
	testutil.AssertEqual(t, humanID, newHostID, "New host should be the human player")

	newHost, _ := updatedGame.GetPlayer(newHostID)
	testutil.AssertFalse(t, newHost.IsBot(), "New host should not be a bot")
}
