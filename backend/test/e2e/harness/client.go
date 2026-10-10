package harness

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"openmars/internal/delivery/dto"

	"github.com/gorilla/websocket"
)

// Message is one server-to-client WebSocket message with its payload still encoded.
type Message struct {
	Type    dto.MessageType `json:"type"`
	GameID  string          `json:"gameId,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// Decode unmarshals the payload into v.
func (m Message) Decode(t testing.TB, v any) {
	t.Helper()
	if err := json.Unmarshal(m.Payload, v); err != nil {
		t.Fatalf("decode %s payload: %v\n%s", m.Type, err, m.Payload)
	}
}

// Client is one browser tab: a WebSocket connection plus everything it has received.
type Client struct {
	Name string

	conn    *websocket.Conn
	arrived chan struct{}
	stopped chan struct{}

	mu       sync.Mutex
	inbox    []Message
	latest   *dto.GameDto
	readErr  error
	writeMu  sync.Mutex
	syncSeq  atomic.Int64
	playerID string
}

// Dial opens a WebSocket connection to the server. Cleanup closes it.
func (s *Server) Dial(t testing.TB, name string) *Client {
	t.Helper()
	c := s.newClient(t, name, s.dialRaw(t, websocket.DefaultDialer))
	go c.read()
	return c
}

// DialStalled opens a connection that never reads anything the server sends, with a
// tiny receive buffer, like a frozen browser tab.
func (s *Server) DialStalled(t testing.TB, name string) *Client {
	t.Helper()
	dialer := *websocket.DefaultDialer
	dialer.NetDialContext = (&net.Dialer{Control: func(_, _ string, raw syscall.RawConn) error {
		var sockErr error
		err := raw.Control(func(fd uintptr) {
			sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, 4096)
		})
		if err != nil {
			return err
		}
		return sockErr
	}}).DialContext
	c := s.newClient(t, name, s.dialRaw(t, &dialer))
	close(c.stopped)
	return c
}

func (s *Server) dialRaw(t testing.TB, dialer *websocket.Dialer) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws"
	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	_ = resp.Body.Close()
	return conn
}

func (s *Server) newClient(t testing.TB, name string, conn *websocket.Conn) *Client {
	c := &Client{
		Name:    name,
		conn:    conn,
		arrived: make(chan struct{}, 1),
		stopped: make(chan struct{}),
	}
	t.Cleanup(c.Close)
	return c
}

func (c *Client) read() {
	defer close(c.stopped)
	for {
		_, raw, err := c.conn.ReadMessage()
		c.mu.Lock()
		if err != nil {
			c.readErr = err
			c.mu.Unlock()
			c.notify()
			return
		}
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.readErr = fmt.Errorf("server sent invalid JSON: %w: %s", err, raw)
			c.mu.Unlock()
			c.notify()
			return
		}
		if msg.Type == dto.MessageTypeGameUpdated {
			var payload dto.GameUpdatedPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				c.latest = &payload.Game
			}
		}
		c.inbox = append(c.inbox, msg)
		c.mu.Unlock()
		c.notify()
	}
}

func (c *Client) notify() {
	select {
	case c.arrived <- struct{}{}:
	default:
	}
}

// Close closes the connection the way a closed browser tab does and waits for the
// reader to stop.
func (c *Client) Close() {
	c.writeMu.Lock()
	_ = c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), time.Now().Add(time.Second))
	c.writeMu.Unlock()
	_ = c.conn.Close()
	<-c.stopped
}

// Closed reports whether the connection has ended.
func (c *Client) Closed() bool {
	select {
	case <-c.stopped:
		return true
	default:
		return false
	}
}

// Send sends a message with the given payload, which is encoded as JSON.
func (c *Client) Send(t testing.TB, messageType dto.MessageType, payload any) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"type": messageType, "payload": payload})
	if err != nil {
		t.Fatalf("encode %s: %v", messageType, err)
	}
	c.SendRaw(t, raw)
}

// TrySend sends a message and reports false instead of failing when the server has
// already closed the connection.
func (c *Client) TrySend(t testing.TB, messageType dto.MessageType, payload any) bool {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"type": messageType, "payload": payload})
	if err != nil {
		t.Fatalf("encode %s: %v", messageType, err)
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(Timeout()))
	return c.conn.WriteMessage(websocket.TextMessage, raw) == nil
}

// SendRaw sends bytes exactly as given, for malformed-frame tests.
func (c *Client) SendRaw(t testing.TB, raw []byte) {
	t.Helper()
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		t.Fatalf("%s: send: %v", c.Name, err)
	}
}

// Await waits for a message of the given type that satisfies match (nil matches any),
// removes it from the inbox and returns it. Other messages stay in the inbox.
func (c *Client) Await(t testing.TB, messageType dto.MessageType, match func(Message) bool) Message {
	t.Helper()
	var found Message
	c.wait(t, fmt.Sprintf("a %s message", messageType), func() bool {
		for i, msg := range c.inbox {
			if msg.Type == messageType && (match == nil || match(msg)) {
				found = msg
				c.inbox = append(c.inbox[:i:i], c.inbox[i+1:]...)
				return true
			}
		}
		return false
	})
	return found
}

// AwaitState waits until the newest game state this client received satisfies match and
// returns it. Older game-updated messages are dropped from the inbox.
func (c *Client) AwaitState(t testing.TB, what string, match func(dto.GameDto) bool) dto.GameDto {
	t.Helper()
	var state dto.GameDto
	c.wait(t, "game state where "+what, func() bool {
		if c.latest == nil || !match(*c.latest) {
			return false
		}
		state = *c.latest
		kept := c.inbox[:0]
		for _, msg := range c.inbox {
			if msg.Type != dto.MessageTypeGameUpdated {
				kept = append(kept, msg)
			}
		}
		c.inbox = kept
		return true
	})
	return state
}

// AwaitError waits for an error message and returns its payload.
func (c *Client) AwaitError(t testing.TB) dto.ErrorPayload {
	t.Helper()
	var payload dto.ErrorPayload
	c.Await(t, dto.MessageTypeError, nil).Decode(t, &payload)
	if payload.Message == "" {
		t.Fatalf("%s: error message has no text", c.Name)
	}
	return payload
}

// AwaitClosed waits until the server closes the connection.
func (c *Client) AwaitClosed(t testing.TB) {
	t.Helper()
	select {
	case <-c.stopped:
	case <-time.After(Timeout()):
		t.Fatalf("%s: server did not close the connection within %s", c.Name, Timeout())
	}
}

// TrySync is Sync for a connection the server may have closed: it reports false instead
// of failing when the connection has ended.
func (c *Client) TrySync(t testing.TB) ([]Message, bool) {
	t.Helper()
	if c.Closed() {
		return nil, false
	}
	requestID := fmt.Sprintf("sync-%s-%d", c.Name, c.syncSeq.Add(1))
	raw, _ := json.Marshal(map[string]any{"type": dto.MessageTypeQuotePayment, "payload": map[string]any{"requestId": requestID}})
	c.writeMu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, raw)
	c.writeMu.Unlock()
	if err != nil {
		return nil, false
	}
	deadline := time.After(Timeout())
	for {
		c.mu.Lock()
		for i, msg := range c.inbox {
			var reply struct {
				RequestID string `json:"requestId"`
			}
			if msg.Type == dto.MessageTypePaymentQuote && json.Unmarshal(msg.Payload, &reply) == nil && reply.RequestID == requestID {
				before := append([]Message(nil), c.inbox[:i]...)
				c.inbox = append([]Message(nil), c.inbox[i+1:]...)
				c.mu.Unlock()
				return before, true
			}
		}
		ended := c.readErr != nil
		c.mu.Unlock()
		if ended {
			return nil, false
		}
		select {
		case <-c.arrived:
		case <-deadline:
			t.Fatalf("%s: timed out after %s waiting for sync reply %s", c.Name, Timeout(), requestID)
		}
	}
}

// Sync is a barrier: it returns once every message the server queued for this client
// before Sync was called has arrived, and returns and clears those messages. The server
// handles messages one at a time on a single goroutine and each connection's queue is
// FIFO, so the reply to this request comes after everything queued earlier.
//
// Messages from different connections are not ordered against each other, so after one
// client acts, Sync that client first (or await its reply) before syncing the others.
func (c *Client) Sync(t testing.TB) []Message {
	t.Helper()
	requestID := fmt.Sprintf("sync-%s-%d", c.Name, c.syncSeq.Add(1))
	c.Send(t, dto.MessageTypeQuotePayment, map[string]any{"requestId": requestID})
	var before []Message
	c.wait(t, "sync reply "+requestID, func() bool {
		for i, msg := range c.inbox {
			if msg.Type != dto.MessageTypePaymentQuote {
				continue
			}
			var reply struct {
				RequestID string `json:"requestId"`
			}
			if json.Unmarshal(msg.Payload, &reply) == nil && reply.RequestID == requestID {
				before = append(before, c.inbox[:i]...)
				c.inbox = append([]Message(nil), c.inbox[i+1:]...)
				return true
			}
		}
		return false
	})
	return before
}

// ExpectQuiet fails the test if any message reached this client since the last Sync,
// Drain or Await that consumed it.
func (c *Client) ExpectQuiet(t testing.TB) {
	t.Helper()
	if got := c.Sync(t); len(got) > 0 {
		types := make([]string, len(got))
		for i, msg := range got {
			types[i] = string(msg.Type)
		}
		t.Fatalf("%s: expected no messages, got %v", c.Name, types)
	}
}

// Drain discards everything received so far, after a Sync.
func (c *Client) Drain(t testing.TB) {
	t.Helper()
	c.Sync(t)
}

// State returns the newest game state this client received.
func (c *Client) State(t testing.TB) dto.GameDto {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.latest == nil {
		t.Fatalf("%s: no game state received yet", c.Name)
	}
	return *c.latest
}

// PlayerID returns the player this client joined as.
func (c *Client) PlayerID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.playerID
}

// Join joins a game as a new player and waits for the first game state. It returns the
// player ID the server assigned.
func (c *Client) Join(t testing.TB, gameID, playerName string) string {
	t.Helper()
	return c.join(t, map[string]any{"gameId": gameID, "playerName": playerName})
}

// Rejoin connects as an existing player, as a returning or second device does.
func (c *Client) Rejoin(t testing.TB, gameID, playerID, playerName string) {
	t.Helper()
	c.join(t, map[string]any{"gameId": gameID, "playerName": playerName, "playerId": playerID})
}

func (c *Client) join(t testing.TB, payload map[string]any) string {
	t.Helper()
	c.Send(t, dto.MessageTypePlayerConnect, payload)
	var connected dto.PlayerConnectedPayload
	c.Await(t, dto.MessageTypePlayerConnected, nil).Decode(t, &connected)
	if connected.PlayerID == "" {
		t.Fatalf("%s: player-connected has no playerId", c.Name)
	}
	c.mu.Lock()
	c.playerID = connected.PlayerID
	c.mu.Unlock()
	c.AwaitState(t, "the joined player is the viewer", func(g dto.GameDto) bool {
		return g.ViewingPlayerID == connected.PlayerID
	})
	return connected.PlayerID
}

// Spectate joins a game as a spectator and returns the spectator ID.
func (c *Client) Spectate(t testing.TB, gameID, name string) string {
	t.Helper()
	c.Send(t, dto.MessageTypeSpectatorConnect, map[string]any{"gameId": gameID, "spectatorName": name})
	var connected dto.SpectatorConnectedPayload
	c.Await(t, dto.MessageTypeSpectatorConnected, nil).Decode(t, &connected)
	if connected.SpectatorID == "" {
		t.Fatalf("%s: spectator-connected has no spectatorId", c.Name)
	}
	c.AwaitState(t, "the spectator view", func(g dto.GameDto) bool { return g.IsSpectator })
	return connected.SpectatorID
}

func (c *Client) wait(t testing.TB, what string, done func() bool) {
	t.Helper()
	deadline := time.After(Timeout())
	for {
		c.mu.Lock()
		ok := done()
		readErr := c.readErr
		inbox := summarize(c.inbox)
		c.mu.Unlock()
		if ok {
			return
		}
		if readErr != nil {
			t.Fatalf("%s: connection ended while waiting for %s: %v (inbox: %s)", c.Name, what, readErr, inbox)
		}
		select {
		case <-c.arrived:
		case <-deadline:
			t.Fatalf("%s: timed out after %s waiting for %s (inbox: %s)", c.Name, Timeout(), what, inbox)
		}
	}
}

func summarize(inbox []Message) string {
	if len(inbox) == 0 {
		return "empty"
	}
	parts := make([]string, len(inbox))
	for i, msg := range inbox {
		parts[i] = string(msg.Type)
		if msg.Type == dto.MessageTypeError {
			parts[i] += string(msg.Payload)
		}
	}
	return strings.Join(parts, ", ")
}
