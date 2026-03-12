package game_lifecycle_test

import (
	"context"
	"testing"

	gameAction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestCreateGameAction_Success(t *testing.T) {
	// Setup
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, logger)

	// Execute
	settings := shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	createdGame, err := createAction.Execute(context.Background(), settings)

	// Assert
	testutil.AssertNoError(t, err, "Failed to create game")
	testutil.AssertNotEqual(t, "", createdGame.ID(), "Game ID should not be empty")
	testutil.AssertEqual(t, shared.GameStatusLobby, createdGame.Status(), "Game should start in lobby status")
	testutil.AssertEqual(t, 4, createdGame.Settings().MaxPlayers, "Max players should be 4")

	// Verify game exists in repository
	fetchedGame, err := repo.Get(context.Background(), createdGame.ID())
	testutil.AssertNoError(t, err, "Failed to fetch created game")
	testutil.AssertEqual(t, createdGame.ID(), fetchedGame.ID(), "Fetched game ID should match")
}

func TestCreateGameAction_DefaultSettings(t *testing.T) {
	// Setup
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, logger)

	// Execute with empty settings
	settings := shared.GameSettings{}
	createdGame, err := createAction.Execute(context.Background(), settings)

	// Assert defaults are applied
	testutil.AssertNoError(t, err, "Failed to create game with empty settings")
	testutil.AssertEqual(t, game.DefaultMaxPlayers, createdGame.Settings().MaxPlayers, "Should use default max players")
	testutil.AssertTrue(t, len(createdGame.Settings().CardPacks) > 0, "Should have default card packs")
}

func TestCreateGameAction_DeckInitialization(t *testing.T) {
	// Setup
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, logger)

	// Execute
	settings := shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	createdGame, err := createAction.Execute(context.Background(), settings)

	// Assert
	testutil.AssertNoError(t, err, "Failed to create game")

	deck := createdGame.Deck()
	testutil.AssertTrue(t, deck != nil, "Deck should be initialized")

	// Verify deck has cards from the base pack
	// (The deck should contain project cards from the registry)
}

func TestCreateGameAction_MultipleCardPacks(t *testing.T) {
	// Setup
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, logger)

	// Execute with multiple packs
	settings := shared.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base", "prelude"},
	}

	createdGame, err := createAction.Execute(context.Background(), settings)

	// Assert
	testutil.AssertNoError(t, err, "Failed to create game with multiple card packs")
	testutil.AssertTrue(t, createdGame.Deck() != nil, "Deck should be initialized with multiple packs")
}

func TestCreateGameAction_BoardInitialization(t *testing.T) {
	// Setup
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, logger)

	// Execute
	settings := shared.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	createdGame, err := createAction.Execute(context.Background(), settings)

	// Assert
	testutil.AssertNoError(t, err, "Failed to create game")

	board := createdGame.Board()
	testutil.AssertTrue(t, board != nil, "Board should be initialized")
}
