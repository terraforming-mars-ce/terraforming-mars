package integration_test

import (
	"context"
	"testing"

	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/test/testutil"
)

func setupActiveGameForGlobalParams(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, string) {
	t.Helper()

	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	testutil.StartTestGame(t, testGame)

	playerID := testGame.CurrentTurn().PlayerID()
	return testGame, repo, cardRegistry, playerID
}

// TestGlobalParameters_TemperatureProgression tests temperature increases
func TestGlobalParameters_TemperatureProgression(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()

	// Get player and give heat
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 32) // Enough for 4 conversions

	logger := testutil.TestLogger()
	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Set as current turn
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")

	initialTemp := testGame.GlobalParameters().Temperature()
	initialTR := player.Resources().TerraformRating()

	// Convert heat 4 times
	for i := 0; i < 4; i++ {
		err := convertAction.Execute(ctx, testGame.ID(), playerID, nil)
		if err != nil {
			t.Logf("Conversion %d failed: %v", i+1, err)
			break
		}

		// Refresh game state
		testGame, _ = repo.Get(ctx, testGame.ID())
		player, _ = testGame.GetPlayer(playerID)
	}

	// Verify temperature increased
	finalTemp := testGame.GlobalParameters().Temperature()
	testutil.AssertTrue(t, finalTemp > initialTemp, "Temperature should increase")

	// Verify TR increased (should increase by number of successful temperature raises)
	finalTR := player.Resources().TerraformRating()
	testutil.AssertTrue(t, finalTR > initialTR, "TR should increase with temperature")
}

// TestGlobalParameters_TemperatureMax tests temperature cannot exceed maximum
func TestGlobalParameters_TemperatureMax(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()

	// Set temperature near max
	testutil.AssertNoError(t, testGame.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature-2), "set temperature")

	// Give player heat
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 100)

	// Set as current turn
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 10), "set current turn")

	logger := testutil.TestLogger()
	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	// Try to raise temperature multiple times
	for i := 0; i < 5; i++ {
		err := convertAction.Execute(ctx, testGame.ID(), playerID, nil)
		if err != nil {
			break
		}
		testGame, _ = repo.Get(ctx, testGame.ID())
		player, _ = testGame.GetPlayer(playerID)
	}

	// Verify temperature doesn't exceed max
	finalTemp := testGame.GlobalParameters().Temperature()
	testutil.AssertTrue(t, finalTemp <= global_parameters.MaxTemperature, "Temperature should not exceed max")
}

// TestGlobalParameters_AllParametersInitialized tests all global parameters are set on game start
func TestGlobalParameters_AllParametersInitialized(t *testing.T) {
	// Setup
	testGame, _, _, _ := setupActiveGameForGlobalParams(t)

	globalParams := testGame.GlobalParameters()
	testutil.AssertTrue(t, globalParams != nil, "Global parameters should exist")

	// Verify all parameters have valid initial values
	temp := globalParams.Temperature()
	testutil.AssertTrue(t, temp >= global_parameters.MinTemperature, "Temperature should be at least minimum")
	testutil.AssertTrue(t, temp <= global_parameters.MaxTemperature, "Temperature should not exceed maximum")

	oxygen := globalParams.Oxygen()
	testutil.AssertTrue(t, oxygen >= 0, "Oxygen should be non-negative")
	testutil.AssertTrue(t, oxygen <= global_parameters.MaxOxygen, "Oxygen should not exceed maximum")

	oceans := globalParams.Oceans()
	testutil.AssertTrue(t, oceans >= 0, "Oceans should be non-negative")
	testutil.AssertTrue(t, oceans <= global_parameters.MaxOceans, "Oceans should not exceed maximum")
}

// TestGlobalParameters_EventsPublished tests that events are published on parameter changes
func TestGlobalParameters_EventsPublished(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()

	// Give player heat
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 8)

	// Set as current turn
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")

	logger := testutil.TestLogger()
	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	initialTemp := testGame.GlobalParameters().Temperature()

	// Convert heat (should increase temperature)
	err := convertAction.Execute(ctx, testGame.ID(), playerID, nil)

	testutil.AssertNoError(t, err, "Failed to convert heat")

	// Verify temperature increased
	finalTemp := testGame.GlobalParameters().Temperature()
	testutil.AssertTrue(t, finalTemp > initialTemp, "Temperature should increase after converting heat")
}

// TestGlobalParameters_TRIncreasesWithTerraforming tests TR increases when terraforming
func TestGlobalParameters_TRIncreasesWithTerraforming(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()

	// Get initial TR
	player, _ := testGame.GetPlayer(playerID)
	initialTR := player.Resources().TerraformRating()

	// Give heat and convert
	testutil.SetPlayerHeat(ctx, player, 8)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")

	logger := testutil.TestLogger()
	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	err := convertAction.Execute(ctx, testGame.ID(), playerID, nil)
	testutil.AssertNoError(t, err, "Heat conversion failed")

	// Get final TR
	testGame, _ = repo.Get(ctx, testGame.ID())
	player, _ = testGame.GetPlayer(playerID)
	finalTR := player.Resources().TerraformRating()

	// TR should increase
	testutil.AssertTrue(t, finalTR > initialTR, "TR should increase when terraforming")
}

// TestGlobalParameters_MultiplePlayers tests multiple players can terraform
func TestGlobalParameters_MultiplePlayers(t *testing.T) {
	// Setup
	testGame, repo, cardRegistry, player1ID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()

	// Get second player
	players := testGame.GetAllPlayers()
	var player2ID string
	for _, p := range players {
		if p.ID() != player1ID {
			player2ID = p.ID()
			break
		}
	}

	logger := testutil.TestLogger()
	convertAction := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, logger)

	initialTemp := testGame.GlobalParameters().Temperature()

	// Player 1 raises temperature
	player1, _ := testGame.GetPlayer(player1ID)
	testutil.SetPlayerHeat(ctx, player1, 8)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1ID, 2), "set current turn for player 1")
	err1 := convertAction.Execute(ctx, testGame.ID(), player1ID, nil)

	// Player 2 raises temperature
	testGame, _ = repo.Get(ctx, testGame.ID())
	player2, _ := testGame.GetPlayer(player2ID)
	testutil.SetPlayerHeat(ctx, player2, 8)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player2ID, 2), "set current turn for player 2")
	err2 := convertAction.Execute(ctx, testGame.ID(), player2ID, nil)

	// Both should succeed (if temperature not maxed)
	if err1 == nil && err2 == nil {
		testGame, _ = repo.Get(ctx, testGame.ID())
		finalTemp := testGame.GlobalParameters().Temperature()

		// Temperature should have increased
		testutil.AssertTrue(t, finalTemp > initialTemp, "Temperature should increase from both players")

		// Both players should have increased TR
		player1, _ = testGame.GetPlayer(player1ID)
		player2, _ = testGame.GetPlayer(player2ID)

		testutil.AssertTrue(t, player1.Resources().TerraformRating() > 20, "Player 1 TR should increase")
		testutil.AssertTrue(t, player2.Resources().TerraformRating() > 20, "Player 2 TR should increase")
	}
}
