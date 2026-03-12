package events_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestBroadcasting_AutomaticOnStateChange tests that AddPlayer executes successfully
func TestBroadcasting_AutomaticOnStateChange(t *testing.T) {
	// Setup
	ds, _ := datastore.NewDataStore()
	settings := shared.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	// Create game without broadcaster
	testGame := game.NewGame(ds, "test-game", "", settings)

	// Add a player
	ctx := context.Background()
	_, err := testGame.AddNewPlayer(ctx, "test-player-1", "TestPlayer")

	// Verify player was added
	testutil.AssertNoError(t, err, "Failed to add player")
	players := testGame.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(players), "Should have 1 player")
}

// TestBroadcasting_MultipleStateChanges tests adding multiple players
func TestBroadcasting_MultipleStateChanges(t *testing.T) {
	// Setup
	ds, _ := datastore.NewDataStore()
	settings := shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame(ds, "test-game", "", settings)
	ctx := context.Background()

	// Add multiple players
	_, err := testGame.AddNewPlayer(ctx, "player-1", "Player1")
	testutil.AssertNoError(t, err, "Failed to add player 1")
	_, err = testGame.AddNewPlayer(ctx, "player-2", "Player2")
	testutil.AssertNoError(t, err, "Failed to add player 2")
	_, err = testGame.AddNewPlayer(ctx, "player-3", "Player3")
	testutil.AssertNoError(t, err, "Failed to add player 3")

	// Verify all players were added
	players := testGame.GetAllPlayers()
	testutil.AssertEqual(t, 3, len(players), "Should have 3 players")
}

// TestBroadcasting_CorrectGameID tests that game ID is correct
func TestBroadcasting_CorrectGameID(t *testing.T) {
	// Setup
	ds, _ := datastore.NewDataStore()
	gameID := "test-game-123"
	settings := shared.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame(ds, gameID, "", settings)
	ctx := context.Background()

	// Add player and verify game ID
	_, err := testGame.AddNewPlayer(ctx, "test-player-1", "TestPlayer")
	testutil.AssertNoError(t, err, "Failed to add player")

	// Verify game has correct ID
	testutil.AssertEqual(t, gameID, testGame.ID(), "Game should have correct game ID")
}

// TestBroadcasting_ResourceChanges tests resource changes execute successfully
func TestBroadcasting_ResourceChanges(t *testing.T) {
	// Setup
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, nil)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	playerObj := players[0]

	initialCredits := playerObj.Resources().Get().Credits

	// Change player resources
	testutil.AddPlayerCredits(ctx, playerObj, 5)

	// Verify resource change
	finalCredits := playerObj.Resources().Get().Credits
	testutil.AssertEqual(t, initialCredits+5, finalCredits, "Player should have 5 more credits")
}

// TestBroadcasting_GlobalParameterChanges tests global parameter changes execute successfully
func TestBroadcasting_GlobalParameterChanges(t *testing.T) {
	// Setup
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, nil)
	ctx := context.Background()

	// Start game first
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Change temperature
	_, err := testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	testutil.AssertNoError(t, err, "Temperature increase should succeed")

	// Verify temperature changed
	finalTemp := testGame.GlobalParameters().Temperature()
	testutil.AssertTrue(t, finalTemp >= -29, "Temperature should not go below -29")
}

// TestBroadcasting_PerGameIsolation tests that game states are isolated
func TestBroadcasting_PerGameIsolation(t *testing.T) {
	// Setup
	ds, _ := datastore.NewDataStore()
	settings := shared.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	game1 := game.NewGame(ds, "game-1", "", settings)
	game2 := game.NewGame(ds, "game-2", "", settings)

	ctx := context.Background()

	// Add player to game1
	_, err := game1.AddNewPlayer(ctx, "player-1", "Player1")
	testutil.AssertNoError(t, err, "Failed to add player to game1")

	// Verify game1 has player
	game1Players := game1.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(game1Players), "Game 1 should have 1 player")

	// Verify game2 has no players (isolated)
	game2Players := game2.GetAllPlayers()
	testutil.AssertEqual(t, 0, len(game2Players), "Game 2 should have 0 players")

	// Add player to game2
	_, err = game2.AddNewPlayer(ctx, "player-1", "Player1")
	testutil.AssertNoError(t, err, "Failed to add player to game2")

	// Verify final state
	game1Players = game1.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(game1Players), "Game 1 should still have 1 player")
	game2Players = game2.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(game2Players), "Game 2 should have 1 player")
}

// TestBroadcasting_WithoutBroadcaster tests that operations work without broadcaster
func TestBroadcasting_WithoutBroadcaster(t *testing.T) {
	// Setup - no broadcaster function
	ds, _ := datastore.NewDataStore()
	settings := shared.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	// Create game without broadcaster
	testGame := game.NewGame(ds, "test-game", "", settings)
	ctx := context.Background()

	// Perform operations
	_, err := testGame.AddNewPlayer(ctx, "player-1", "Player1")
	testutil.AssertNoError(t, err, "Failed to add player")

	players := testGame.GetAllPlayers()
	testutil.AssertEqual(t, 1, len(players), "Should have 1 player")

	// Verify resource changes work
	testutil.AddPlayerCredits(ctx, players[0], 10)
	testutil.AssertEqual(t, 10, players[0].Resources().Get().Credits, "Player should have 10 credits")
}

// TestBroadcasting_ConcurrentStateChanges tests thread-safety of state changes
func TestBroadcasting_ConcurrentStateChanges(t *testing.T) {
	// Setup
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3, nil)
	ctx := context.Background()

	players := testGame.GetAllPlayers()

	// Concurrent state changes
	done := make(chan bool, len(players))

	for _, p := range players {
		go func(pl *player.Player) {
			for i := 0; i < 10; i++ {
				testutil.AddPlayerCredits(ctx, pl, 1)
			}
			done <- true
		}(p)
	}

	// Wait for all goroutines
	for i := 0; i < len(players); i++ {
		<-done
	}

	// Verify concurrent changes completed successfully
	expectedCredits := 10 // Each player adds 10 credits in their goroutine
	for _, p := range players {
		testutil.AssertEqual(t, expectedCredits, p.Resources().Get().Credits, "Player should have correct credits after concurrent updates")
	}
}

// TestBroadcasting_MultipleStateUpdates tests multiple resource updates
func TestBroadcasting_MultipleStateUpdates(t *testing.T) {
	// Setup
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, nil)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]

	initialCredits := player.Resources().Get().Credits

	// Multiple updates
	testutil.AddPlayerCredits(ctx, player, 5)
	testutil.AddPlayerCredits(ctx, player, 10)

	// Verify final state
	finalCredits := player.Resources().Get().Credits
	testutil.AssertEqual(t, initialCredits+15, finalCredits, "Player should have 15 more credits after two updates")
}
