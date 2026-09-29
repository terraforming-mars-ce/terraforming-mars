package websocket_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	ws "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connection"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
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

func TestLogUpdates_DistinguishHistoryFromLiveEvents(t *testing.T) {
	g, repo := testutil.CreateTestGameWithPlayers(t, 1, nil)
	playerID := g.GetAllPlayers()[0].ID()
	stateRepo := game.NewInMemoryGameStateRepository()
	entry, err := stateRepo.Write(context.Background(), g.ID(), g, "Played card", shared.SourceTypeCardPlay, playerID, "")
	testutil.AssertNoError(t, err, "record log entry")
	hub := core.NewHub()
	playerConn := core.NewConnection("player-connection", nil, hub.GetManager(), nil, nil)
	playerConn.PlayerID = playerID
	playerConn.GameID = g.ID()
	hub.RegisterConnectionWithGame(playerConn, g.ID())
	spectatorConn := core.NewConnection("spectator-connection", nil, hub.GetManager(), nil, nil)
	spectatorConn.ConnType = core.ConnectionTypeSpectator
	spectatorConn.SpectatorID = "spectator"
	spectatorConn.GameID = g.ID()
	hub.RegisterConnectionWithGame(spectatorConn, g.ID())
	broadcaster := ws.NewBroadcaster(repo, stateRepo, hub, testutil.CreateTestCardRegistry(), nil, nil, nil, nil, nil, nil)

	readLog := func(conn *core.Connection, isHistory bool) {
		t.Helper()
		for {
			select {
			case message := <-conn.Send:
				if message.Type != dto.MessageTypeLogUpdate {
					continue
				}
				payload, ok := message.Payload.(dto.LogUpdatePayload)
				if !ok {
					t.Fatalf("unexpected log payload %T", message.Payload)
				}
				testutil.AssertEqual(t, isHistory, payload.IsHistory, "history marker")
				testutil.AssertEqual(t, 1, len(payload.Logs), "log entry count")
				testutil.AssertEqual(t, entry.SequenceNumber, payload.Logs[0].SequenceNumber, "log sequence")
				return
			default:
				t.Fatal("expected log update")
			}
		}
	}

	for i := 0; i < 2; i++ {
		broadcaster.SendInitialLogs(g.ID(), playerID)
		readLog(playerConn, true)
		broadcaster.SendInitialLogsToSpectator(g.ID(), spectatorConn.SpectatorID)
		readLog(spectatorConn, true)
	}

	broadcaster.BroadcastGameState(g.ID(), nil)
	readLog(playerConn, false)
	readLog(spectatorConn, false)
	broadcaster.BroadcastLogUpdate(g.ID(), entry)
	readLog(playerConn, false)
	readLog(spectatorConn, false)
}
