package game_lifecycle_test

import (
	"context"
	"testing"

	gameAction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

func TestAddBot_Success(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	testGame.UpdateSettings(context.Background(), game.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		ClaudeModel:  "sonnet",
		CardPacks:    []string{"base-game"},
	})

	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	action := gameAction.NewAddBotAction(repo, cardRegistry, nil, nil, logger)

	result, err := action.Execute(context.Background(), testGame.ID(), "", "", "")
	testutil.AssertNoError(t, err, "Add bot should succeed")
	testutil.AssertTrue(t, result.PlayerID != "", "Bot player ID should not be empty")

	// Verify bot is in game
	players := testGame.GetAllPlayers()
	testutil.AssertEqual(t, 2, len(players), "Should have 2 players")

	// Find the bot player
	botPlayer, err := testGame.GetPlayer(result.PlayerID)
	testutil.AssertNoError(t, err, "Should find bot player")
	testutil.AssertTrue(t, botPlayer.IsBot(), "Player should be a bot")
	testutil.AssertTrue(t, botPlayer.Name() != "", "Bot should have a name")
}

func TestAddBot_NoAPIKey(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	_ = testGame

	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	action := gameAction.NewAddBotAction(repo, cardRegistry, nil, nil, logger)

	_, err := action.Execute(context.Background(), testGame.ID(), "", "", "")
	testutil.AssertError(t, err, "Add bot should fail without API key")
}

func TestAddBot_GameNotInLobby(t *testing.T) {
	g, repo, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	g.UpdateSettings(context.Background(), game.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	logger := testutil.TestLogger()
	action := gameAction.NewAddBotAction(repo, cardRegistry, nil, nil, logger)

	_, err := action.Execute(context.Background(), g.ID(), "", "", "")
	testutil.AssertError(t, err, "Add bot should fail when game is not in lobby")
}

func TestAddBot_GameFull(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 4, broadcaster)
	testGame.UpdateSettings(context.Background(), game.GameSettings{
		MaxPlayers:   4,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	action := gameAction.NewAddBotAction(repo, cardRegistry, nil, nil, logger)

	_, err := action.Execute(context.Background(), testGame.ID(), "", "", "")
	testutil.AssertError(t, err, "Add bot should fail when game is full")
}

func TestAddBot_UniqueNames(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	testGame.UpdateSettings(context.Background(), game.GameSettings{
		MaxPlayers:   5,
		ClaudeAPIKey: "test-key",
		CardPacks:    []string{"base-game"},
	})

	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	action := gameAction.NewAddBotAction(repo, cardRegistry, nil, nil, logger)

	// Add 3 bots
	names := make(map[string]bool)
	for i := 0; i < 3; i++ {
		result, err := action.Execute(context.Background(), testGame.ID(), "", "", "")
		testutil.AssertNoError(t, err, "Bot should be added successfully")
		bot, _ := testGame.GetPlayer(result.PlayerID)
		testutil.AssertTrue(t, !names[bot.Name()], "Bot names should be unique")
		names[bot.Name()] = true
	}
}
