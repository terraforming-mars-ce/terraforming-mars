package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"

	"github.com/gorilla/websocket"
)

type handlerFunc func(ctx context.Context, c *core.Connection, m dto.WebSocketMessage)

func (f handlerFunc) HandleMessage(ctx context.Context, c *core.Connection, m dto.WebSocketMessage) {
	f(ctx, c, m)
}

func serve(t *testing.T, hub *core.Hub) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		hub.Run(ctx)
	}()
	srv := httptest.NewServer(http.HandlerFunc(core.NewHandler(hub).ServeWS))
	t.Cleanup(func() {
		cancel()
		<-done
		srv.Close()
	})
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_ = resp.Body.Close()
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func read(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg map[string]any
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	return msg
}

func TestHub_APanickingHandlerAnswersWithAnErrorAndKeepsServing(t *testing.T) {
	hub := core.NewHub()
	hub.RegisterHandler("explode", handlerFunc(func(context.Context, *core.Connection, dto.WebSocketMessage) {
		var m map[string]int
		m["boom"] = 1
	}))
	hub.RegisterHandler("echo", handlerFunc(func(_ context.Context, c *core.Connection, m dto.WebSocketMessage) {
		c.Send(dto.WebSocketMessage{Type: "echoed", Payload: m.Payload})
	}))
	conn := serve(t, hub)

	if err := conn.WriteJSON(map[string]any{"type": "explode", "payload": map[string]any{}}); err != nil {
		t.Fatalf("send: %v", err)
	}
	reply := read(t, conn)
	payload, _ := reply["payload"].(map[string]any)
	if reply["type"] != string(dto.MessageTypeError) || payload["requestType"] != "explode" {
		t.Fatalf("a panicking handler should answer with an error for its request, got %v", reply)
	}

	if err := conn.WriteJSON(map[string]any{"type": "echo", "payload": "still alive"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if reply := read(t, conn); reply["type"] != "echoed" {
		t.Fatalf("the connection should keep working after a handler panic, got %v", reply)
	}
}

func TestHub_DoRunsOnTheHubAndRecoversPanics(t *testing.T) {
	hub := core.NewHub()
	serve(t, hub)
	ran := false
	if err := hub.Do(context.Background(), func() { ran = true }); err != nil || !ran {
		t.Fatalf("Do should run the job: ran=%v err=%v", ran, err)
	}
	if err := hub.Do(context.Background(), func() { panic("job failed") }); err != nil {
		t.Fatalf("a panicking job should not fail Do: %v", err)
	}
	if err := hub.Do(context.Background(), func() {}); err != nil {
		t.Fatalf("the hub should keep running jobs after a panic: %v", err)
	}
}

func TestHub_DoAfterShutdownReturnsInsteadOfBlocking(t *testing.T) {
	hub := core.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		hub.Run(ctx)
	}()
	cancel()
	<-done
	if err := hub.Do(context.Background(), func() { t.Error("job ran after shutdown") }); err == nil {
		t.Fatalf("Do after shutdown should report that the hub stopped")
	}
}
