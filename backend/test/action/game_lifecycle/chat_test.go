package game_lifecycle_test

import (
	"strings"
	"testing"

	"terraforming-mars-backend/internal/action/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func newSendChatMessageAction(repo game.GameRepository) *connection.SendChatMessageAction {
	return connection.NewSendChatMessageAction(repo, testutil.TestLogger())
}

// ============================================================================
// SendChatMessage - Green Path
// ============================================================================

func TestSendChatMessage_PlayerSendsMessage(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSendChatMessageAction(repo)
	msg, err := action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", "Hello world", false)

	testutil.AssertNoError(t, err, "Player should send chat message")
	testutil.AssertEqual(t, "Player A", msg.SenderName, "Sender name mismatch")
	testutil.AssertEqual(t, "#ff0000", msg.SenderColor, "Sender color mismatch")
	testutil.AssertEqual(t, "Hello world", msg.Message, "Message mismatch")
	testutil.AssertFalse(t, msg.IsSpectator, "Should not be marked as spectator")
}

func TestSendChatMessage_SpectatorSendsMessage(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	spectateAction := newSpectateAction(repo)
	_, _ = spectateAction.Execute(ctx, g.ID(), "Spec1", "spec-1")

	action := newSendChatMessageAction(repo)
	msg, err := action.Execute(ctx, g.ID(), "spec-1", "Spec1", "#9b9b9b", "I'm watching!", true)

	testutil.AssertNoError(t, err, "Spectator should send chat message")
	testutil.AssertEqual(t, "Spec1", msg.SenderName, "Sender name mismatch")
	testutil.AssertTrue(t, msg.IsSpectator, "Should be marked as spectator")
}

func TestSendChatMessage_StoredInGameState(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSendChatMessageAction(repo)
	_, _ = action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", "First", false)
	_, _ = action.Execute(ctx, g.ID(), "player-b", "Player B", "#00ff00", "Second", false)
	_, _ = action.Execute(ctx, g.ID(), "spec-1", "Spec1", "#9b9b9b", "Third", true)

	messages := g.GetChatMessages()
	testutil.AssertEqual(t, 3, len(messages), "Should have 3 chat messages")
	testutil.AssertEqual(t, "First", messages[0].Message, "First message mismatch")
	testutil.AssertEqual(t, "Second", messages[1].Message, "Second message mismatch")
	testutil.AssertEqual(t, "Third", messages[2].Message, "Third message mismatch")
	testutil.AssertTrue(t, messages[2].IsSpectator, "Third message should be from spectator")
}

func TestSendChatMessage_TimestampSet(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSendChatMessageAction(repo)
	msg, err := action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", "Hello", false)

	testutil.AssertNoError(t, err, "Should send message")
	testutil.AssertTrue(t, !msg.Timestamp.IsZero(), "Timestamp should be set")

	messages := g.GetChatMessages()
	testutil.AssertTrue(t, !messages[0].Timestamp.IsZero(), "Stored timestamp should be set")
}

// ============================================================================
// SendChatMessage - Red Path
// ============================================================================

func TestSendChatMessage_EmptyMessage(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSendChatMessageAction(repo)
	_, err := action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", "", false)

	testutil.AssertError(t, err, "Should reject empty message")
}

func TestSendChatMessage_ExceedsMaxLength(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	longMessage := strings.Repeat("a", shared.MaxChatMessageLength+1)
	action := newSendChatMessageAction(repo)
	_, err := action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", longMessage, false)

	testutil.AssertError(t, err, "Should reject message exceeding max length")
}

func TestSendChatMessage_ExactMaxLength(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	exactMessage := strings.Repeat("a", shared.MaxChatMessageLength)
	action := newSendChatMessageAction(repo)
	_, err := action.Execute(ctx, g.ID(), "player-a", "Player A", "#ff0000", exactMessage, false)

	testutil.AssertNoError(t, err, "Should accept message at exact max length")
}

func TestSendChatMessage_GameNotFound(t *testing.T) {
	repo := testutil.NewTestGameRepository(t)
	action := newSendChatMessageAction(repo)

	_, err := action.Execute(testutil.TestContext(), "nonexistent", "player-1", "Player", "#ff0000", "Hello", false)
	testutil.AssertError(t, err, "Should fail for nonexistent game")
}

// ============================================================================
// Chat Message Trimming
// ============================================================================

func TestChatMessage_TrimsOldMessages(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := testutil.TestContext()

	action := newSendChatMessageAction(repo)
	for i := 0; i < shared.MaxChatMessages+10; i++ {
		_, err := action.Execute(ctx, g.ID(), "player-1", "Player", "#ff0000", "msg", false)
		testutil.AssertNoError(t, err, "Should send message")
	}

	messages := g.GetChatMessages()
	testutil.AssertEqual(t, shared.MaxChatMessages, len(messages), "Should trim to MaxChatMessages")
}
