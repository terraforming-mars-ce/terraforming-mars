package websocket_test

import (
	"context"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/confirmation"
	"testing"
)

func TestConfirmResourceRemoval_RejectsMalformedAmount(t *testing.T) {
	for _, amount := range []any{1.5, "4", nil, 1e30} {
		handler := confirmation.NewConfirmResourceRemovalHandler(nil, &mockLogBroadcaster{})
		connection := &core.Connection{GameID: "game", PlayerID: "player", Send: make(chan dto.WebSocketMessage, 1)}
		handler.HandleMessage(context.Background(), connection, dto.WebSocketMessage{Payload: map[string]interface{}{"selectionId": "pending", "targetPlayerId": "target", "amount": amount}})
		select {
		case response := <-connection.Send:
			if response.Type != dto.MessageTypeError {
				t.Fatal("malformed request must fail before executing action")
			}
			payload := response.Payload.(map[string]interface{})
			if payload["selectionId"] != "pending" {
				t.Fatal("error must identify the selection so UI can allow retry")
			}
		default:
			t.Fatal("expected validation error")
		}
	}
}
