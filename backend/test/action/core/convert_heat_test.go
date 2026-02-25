package core_test

import (
	"context"
	"testing"

	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

func setupActiveGame(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, string) {
	t.Helper()

	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	// Start game
	testutil.StartTestGame(t, testGame)

	// Get the current turn player ID (set by StartTestGame)
	playerID := testGame.CurrentTurn().PlayerID()

	return testGame, repo, cardRegistry, playerID
}

func TestConvertHeatAction_Success(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGame(t)
	logger := testutil.TestLogger()

	// Give player enough heat
	ctx := context.Background()
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 8)

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Get initial temperature
	initialTemp := testGame.GlobalParameters().Temperature()

	// Execute
	err := convertAction.Execute(context.Background(), testGame.ID(), playerID)

	// Assert
	testutil.AssertNoError(t, err, "Failed to convert heat")

	fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
	fetchedPlayer, _ := fetchedGame.GetPlayer(playerID)

	// Heat should be reduced by 8
	testutil.AssertEqual(t, 0, testutil.GetPlayerHeat(fetchedPlayer), "Heat should be 0 after conversion")

	// Temperature should increase
	newTemp := fetchedGame.GlobalParameters().Temperature()
	testutil.AssertEqual(t, initialTemp+2, newTemp, "Temperature should increase by 2")

	// TR should increase
	testutil.AssertTrue(t, fetchedPlayer.Resources().TerraformRating() > 20, "TR should increase")
}

func TestConvertHeatAction_InsufficientHeat(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGame(t)
	logger := testutil.TestLogger()

	// Give player insufficient heat
	ctx := context.Background()
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 5)

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Execute
	err := convertAction.Execute(context.Background(), testGame.ID(), playerID)

	// Assert
	testutil.AssertError(t, err, "Should fail with insufficient heat")
}

func TestConvertHeatAction_GameNotFound(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Execute
	err := convertAction.Execute(context.Background(), "non-existent", "player-id")

	// Assert
	testutil.AssertError(t, err, "Should fail when game not found")
}

func TestConvertHeatAction_PlayerNotFound(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, _ := setupActiveGame(t)
	logger := testutil.TestLogger()

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Execute
	err := convertAction.Execute(context.Background(), testGame.ID(), "non-existent-player")

	// Assert
	testutil.AssertError(t, err, "Should fail when player not found")
}

func TestConvertHeatAction_TemperatureMaxed(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGame(t)
	logger := testutil.TestLogger()

	// Set temperature to max
	ctx := context.Background()
	maxTemp := 8
	testGame.GlobalParameters().SetTemperature(ctx, maxTemp)

	// Give player heat
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 8)

	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Execute - should fail or not increase temperature
	err := convertAction.Execute(context.Background(), testGame.ID(), playerID)

	if err == nil {
		// If no error, verify temperature didn't exceed max
		fetchedGame, _ := repo.Get(context.Background(), testGame.ID())
		temp := fetchedGame.GlobalParameters().Temperature()
		testutil.AssertTrue(t, temp <= maxTemp, "Temperature should not exceed max")
	}
}
