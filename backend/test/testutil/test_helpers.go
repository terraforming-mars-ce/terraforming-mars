package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// TestContext provides a reusable test context
func TestContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "test", true)
}

// TestLogger creates a test logger (no-op or minimal output)
func TestLogger() *zap.Logger {
	return logger.Get()
}

// MockBroadcaster records broadcast calls for test assertions.
type MockBroadcaster struct {
	BroadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	GameID    string
	PlayerIDs []string
	Timestamp time.Time
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		BroadcastCalls: make([]BroadcastCall, 0),
	}
}

func (m *MockBroadcaster) CallCount() int {
	return len(m.BroadcastCalls)
}

func (m *MockBroadcaster) Reset() {
	m.BroadcastCalls = make([]BroadcastCall, 0)
}

// CreateTestCardRegistry returns the real card registry loaded from the JSON database.
func CreateTestCardRegistry() cards.CardRegistry {
	return GetCardDB()
}

// CreateTestCardRegistryWithAdditionalCards creates a card registry with real cards plus additional synthetic cards.
func CreateTestCardRegistryWithAdditionalCards(additionalCards []gamecards.Card) cards.CardRegistry {
	allCards := GetCardDB().GetAll()
	allCards = append(allCards, additionalCards...)
	return cards.NewInMemoryCardRegistry(allCards)
}

// CreateTestGameWithPlayers creates a game with specified number of players
func CreateTestGameWithPlayers(t *testing.T, numPlayers int, broadcaster *MockBroadcaster) (*game.Game, game.GameRepository) {
	t.Helper()

	repo := game.NewInMemoryGameRepository()
	cardRegistry := CreateTestCardRegistry()

	// Create game
	settings := game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base-game"},
	}

	testGame := game.NewGame("test-game-id", "", settings)
	allCards := cardRegistry.GetAll()

	// Separate cards by type
	projectCards := make([]string, 0)
	corpCards := make([]string, 0)
	preludeCards := make([]string, 0)

	for _, card := range allCards {
		switch card.Type {
		case gamecards.CardTypeCorporation:
			corpCards = append(corpCards, card.ID)
		case gamecards.CardTypePrelude:
			preludeCards = append(preludeCards, card.ID)
		default:
			projectCards = append(projectCards, card.ID)
		}
	}

	// Create and set deck
	gameDeck := deck.NewDeck(testGame.ID(), projectCards, corpCards, preludeCards)
	testGame.SetDeck(gameDeck)
	testGame.SetVPCardLookup(cards.NewVPCardLookupAdapter(cardRegistry))

	err := repo.Create(context.Background(), testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Add players
	ctx := context.Background()
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("player-%d", i+1)
		playerName := "Player " + string(rune('A'+i))

		// Create player
		newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), playerID, playerName)

		// Add to game
		err := testGame.AddPlayer(ctx, newPlayer)
		if err != nil {
			t.Fatalf("Failed to add player %d: %v", i, err)
		}
	}

	// Set first player as host (like JoinGameAction does)
	if numPlayers > 0 {
		if err := testGame.SetHostPlayerID(ctx, "player-1"); err != nil {
			t.Fatalf("Failed to set host player: %v", err)
		}
	}

	return testGame, repo
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", message)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected == actual {
		t.Fatalf("%s: expected not equal to %v", message, expected)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true, got false", message)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false, got true", message)
	}
}

// IntPtr returns a pointer to the given int value.
func IntPtr(v int) *int { return &v }

// StrPtr returns a pointer to the given string value.
func StrPtr(v string) *string { return &v }

// TagPtr returns a pointer to the given CardTag value.
func TagPtr(v shared.CardTag) *shared.CardTag { return &v }

// ResourceTypePtr returns a pointer to the given ResourceType value.
func ResourceTypePtr(v shared.ResourceType) *shared.ResourceType { return &v }
