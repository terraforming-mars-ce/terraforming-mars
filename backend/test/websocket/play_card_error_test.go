package websocket_test

import (
	"context"
	"encoding/json"
	"testing"

	cardaction "openmars/internal/action/card"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/delivery/websocket/handler/card"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

func TestPlayCardError_IdentifiesRejectedRequest(t *testing.T) {
	for _, tc := range []struct {
		name    string
		gameID  string
		payload any
		cardID  string
	}{
		{"disconnected", "", map[string]any{"cardId": "card-a"}, "card-a"},
		{"invalid payload", "game", "invalid", ""},
		{"missing card", "game", map[string]any{}, ""},
		{"invalid payment", "game", map[string]any{"cardId": "card-b", "payment": "invalid"}, "card-b"},
		{"invalid amount", "game", map[string]any{"cardId": "card-c", "payment": map[string]any{}, "selectedAmount": 1.5}, "card-c"},
		{"invalid storage source", "game", map[string]any{"cardId": "card-d", "payment": map[string]any{}, "cardStorageSources": []any{123}}, "card-d"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := card.NewPlayCardHandler(nil, &mockLogBroadcaster{})
			connection := &core.Connection{GameID: tc.gameID, PlayerID: "player", Send: make(chan dto.WebSocketMessage, 1)}
			handler.HandleMessage(context.Background(), connection, dto.WebSocketMessage{Payload: tc.payload})
			assertPlayCardError(t, connection, tc.cardID)
		})
	}
}

func TestPlayCardError_ExecutionFailureKeepsCardInHand(t *testing.T) {
	ctx := context.Background()
	game, repo := testutil.CreateTestGameWithPlayers(t, 1, testutil.NewMockBroadcaster())
	player := game.GetAllPlayers()[0]
	cardID := testutil.CardID("Space Station")
	player.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player.Hand().AddCard(cardID)
	testutil.AssertNoError(t, game.UpdateStatus(ctx, shared.GameStatusActive), "activate game")
	testutil.AssertNoError(t, game.UpdatePhase(ctx, shared.GamePhaseAction), "action phase")
	testutil.AssertNoError(t, game.SetCurrentTurn(ctx, player.ID(), 2), "set turn")
	action := cardaction.NewPlayCardAction(repo, testutil.CreateTestCardRegistry(), nil, testutil.TestLogger())
	handler := card.NewPlayCardHandler(action, &mockLogBroadcaster{})
	connection := &core.Connection{GameID: game.ID(), PlayerID: player.ID(), Send: make(chan dto.WebSocketMessage, 1)}
	handler.HandleMessage(ctx, connection, dto.WebSocketMessage{Payload: map[string]any{
		"cardId": cardID, "payment": map[string]any{"allocations": []any{}},
	}})
	assertPlayCardError(t, connection, cardID)
	if !player.Hand().HasCard(cardID) {
		t.Fatal("a rejected play must leave the card in the hand")
	}
}

func assertPlayCardError(t *testing.T, connection *core.Connection, cardID string) {
	t.Helper()
	select {
	case response := <-connection.Send:
		if response.Type != dto.MessageTypeError {
			t.Fatalf("expected error, got %s", response.Type)
		}
		encoded, err := json.Marshal(response.Payload)
		if err != nil {
			t.Fatal(err)
		}
		var payload dto.PlayCardErrorPayload
		if err := json.Unmarshal(encoded, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Action != "play-card" || payload.CardID != cardID || payload.Error == "" {
			t.Fatalf("error must identify the failed play and its reason: %s", encoded)
		}
	default:
		t.Fatal("expected a correlated play-card error")
	}
}
