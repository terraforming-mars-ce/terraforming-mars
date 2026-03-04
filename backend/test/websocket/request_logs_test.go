package websocket_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connection"
	"terraforming-mars-backend/test/testutil"
)

type mockLogBroadcaster struct {
	sendInitialLogsCalls []sendInitialLogsCall
}

type sendInitialLogsCall struct {
	GameID   string
	PlayerID string
}

func (m *mockLogBroadcaster) BroadcastGameState(_ string, _ []string) {}

func (m *mockLogBroadcaster) SendInitialLogs(gameID string, playerID string) {
	m.sendInitialLogsCalls = append(m.sendInitialLogsCalls, sendInitialLogsCall{
		GameID:   gameID,
		PlayerID: playerID,
	})
}

func (m *mockLogBroadcaster) SendInitialLogsToSpectator(_ string, _ string) {}

func TestRequestLogsHandler_SendsInitialLogs(t *testing.T) {
	broadcaster := &mockLogBroadcaster{}
	handler := connection.NewRequestLogsHandler(broadcaster)

	conn := &core.Connection{
		ID:       "conn-1",
		GameID:   "game-123",
		PlayerID: "player-456",
		Send:     make(chan dto.WebSocketMessage, 10),
	}

	message := dto.WebSocketMessage{
		Type: dto.MessageTypeRequestLogs,
	}

	handler.HandleMessage(context.Background(), conn, message)

	testutil.AssertEqual(t, 1, len(broadcaster.sendInitialLogsCalls), "Should call SendInitialLogs once")
	testutil.AssertEqual(t, "game-123", broadcaster.sendInitialLogsCalls[0].GameID, "Should pass correct game ID")
	testutil.AssertEqual(t, "player-456", broadcaster.sendInitialLogsCalls[0].PlayerID, "Should pass correct player ID")
}

func TestRequestLogsHandler_MissingConnectionContext(t *testing.T) {
	broadcaster := &mockLogBroadcaster{}
	handler := connection.NewRequestLogsHandler(broadcaster)

	conn := &core.Connection{
		ID:   "conn-1",
		Send: make(chan dto.WebSocketMessage, 10),
	}

	message := dto.WebSocketMessage{
		Type: dto.MessageTypeRequestLogs,
	}

	handler.HandleMessage(context.Background(), conn, message)

	testutil.AssertEqual(t, 0, len(broadcaster.sendInitialLogsCalls), "Should not call SendInitialLogs")

	select {
	case msg := <-conn.Send:
		testutil.AssertEqual(t, dto.MessageTypeError, msg.Type, "Should send error message")
	default:
		t.Fatal("Expected error message on send channel")
	}
}
