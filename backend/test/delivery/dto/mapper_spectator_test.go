package dto_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestToSpectatorGameDto_AllPlayersAsOtherPlayers(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	gameDto := dto.ToSpectatorGameDto(testGame, cardRegistry)

	testutil.AssertEqual(t, 3, len(gameDto.OtherPlayers), "All players should be OtherPlayers")
	testutil.AssertEqual(t, "", gameDto.CurrentPlayer.ID, "CurrentPlayer should be empty")
	testutil.AssertEqual(t, "", gameDto.ViewingPlayerID, "ViewingPlayerID should be empty")
}

func TestToSpectatorGameDto_IsSpectatorTrue(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	gameDto := dto.ToSpectatorGameDto(testGame, cardRegistry)

	testutil.AssertTrue(t, gameDto.IsSpectator, "IsSpectator should be true")
}

func TestToSpectatorGameDto_NoHandCards(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	gameDto := dto.ToSpectatorGameDto(testGame, cardRegistry)

	for _, op := range gameDto.OtherPlayers {
		testutil.AssertEqual(t, 0, op.HandCardCount, "OtherPlayer hand card count should be 0 initially")
	}
}

func TestToSpectatorGameDto_IncludesSpectators(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	spec := game.NewSpectator("spec-1", "Watcher", "#9b9b9b")
	testutil.AssertNoError(t, testGame.AddSpectator(ctx, spec), "add spectator")

	gameDto := dto.ToSpectatorGameDto(testGame, cardRegistry)

	testutil.AssertEqual(t, 1, len(gameDto.Spectators), "Should include spectators")
	testutil.AssertEqual(t, "spec-1", gameDto.Spectators[0].ID, "Spectator ID mismatch")
	testutil.AssertEqual(t, "Watcher", gameDto.Spectators[0].Name, "Spectator name mismatch")
	testutil.AssertEqual(t, "#9b9b9b", gameDto.Spectators[0].Color, "Spectator color mismatch")
}

func TestToSpectatorGameDto_IncludesChatMessages(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	testGame.AddChatMessage(ctx, shared.ChatMessage{
		SenderName:  "Player A",
		SenderColor: "#ff0000",
		Message:     "Hello",
		Timestamp:   time.Now(),
		IsSpectator: false,
	})
	testGame.AddChatMessage(ctx, shared.ChatMessage{
		SenderName:  "Spectator",
		SenderColor: "#9b9b9b",
		Message:     "Hi there",
		Timestamp:   time.Now(),
		IsSpectator: true,
	})

	gameDto := dto.ToSpectatorGameDto(testGame, cardRegistry)

	testutil.AssertEqual(t, 2, len(gameDto.ChatMessages), "Should include chat messages")
	testutil.AssertEqual(t, "Hello", gameDto.ChatMessages[0].Message, "First message mismatch")
	testutil.AssertTrue(t, gameDto.ChatMessages[1].IsSpectator, "Second message should be from spectator")
}

func TestToGameDto_IncludesSpectatorList(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	spec := game.NewSpectator("spec-1", "Watcher", "#9b9b9b")
	testutil.AssertNoError(t, testGame.AddSpectator(ctx, spec), "add spectator")

	players := testGame.GetAllPlayers()
	gameDto := dto.ToGameDto(testGame, cardRegistry, players[0].ID())

	testutil.AssertEqual(t, 1, len(gameDto.Spectators), "Regular GameDto should include spectators")
	testutil.AssertFalse(t, gameDto.IsSpectator, "Regular GameDto should not be marked as spectator")
}

func TestToGameDto_IncludesChatMessages(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	testGame.AddChatMessage(ctx, shared.ChatMessage{
		SenderName:  "Player A",
		SenderColor: "#ff0000",
		Message:     "Test message",
		Timestamp:   time.Now(),
		IsSpectator: false,
	})

	players := testGame.GetAllPlayers()
	gameDto := dto.ToGameDto(testGame, cardRegistry, players[0].ID())

	testutil.AssertEqual(t, 1, len(gameDto.ChatMessages), "Regular GameDto should include chat messages")
	testutil.AssertEqual(t, "Test message", gameDto.ChatMessages[0].Message, "Chat message mismatch")
}
