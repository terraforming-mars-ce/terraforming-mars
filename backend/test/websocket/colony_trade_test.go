package websocket_test

import (
	"context"
	"fmt"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	colonyhandler "openmars/internal/delivery/websocket/handler/colony"
	"openmars/internal/delivery/websocket/handler/confirmation"
	"testing"
)

func TestTradeRequests_RejectMissingOrMalformedTrackChoice(t *testing.T) {
	for _, free := range []bool{false, true} {
		for _, value := range []any{nil, "1", 1.5, 1e30, true, "missing"} {
			t.Run(fmt.Sprintf("free%v/%v", free, value), func(t *testing.T) {
				connection := &core.Connection{GameID: "game", PlayerID: "player", Send: make(chan dto.WebSocketMessage, 1)}
				payload := map[string]interface{}{"colonyId": "luna", "paymentType": "energy"}
				if value != "missing" {
					payload["trackSteps"] = value
				}
				message := dto.WebSocketMessage{Payload: payload}
				if free {
					confirmation.NewConfirmFreeTradeHandler(nil, &mockLogBroadcaster{}).HandleMessage(context.Background(), connection, message)
				} else {
					colonyhandler.NewTradeHandler(nil, &mockLogBroadcaster{}).HandleMessage(context.Background(), connection, message)
				}
				select {
				case response := <-connection.Send:
					if response.Type != dto.MessageTypeError {
						t.Fatal("malformed choice must fail before executing action")
					}
				default:
					t.Fatal("expected validation error")
				}
			})
		}
	}
}
