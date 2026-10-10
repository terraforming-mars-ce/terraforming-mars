// Package harness runs the real server in-process and drives it the way a browser does:
// over HTTP and real WebSocket connections. Waits are always "until a condition holds or
// the timeout passes", never sleeps.
package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"openmars/internal/app"
	"openmars/internal/delivery/dto"
	"openmars/internal/service/bot"
	"openmars/test/testutil"
)

// Timeout bounds every wait in the harness. The race detector and coverage
// instrumentation slow the server down by an order of magnitude, so those runs wait longer.
func Timeout() time.Duration {
	if Race || testing.CoverMode() != "" {
		return 60 * time.Second
	}
	return 5 * time.Second
}

// Server is a running server with production wiring.
type Server struct {
	URL string

	app       *app.App
	http      *httptest.Server
	closeOnce sync.Once
}

type options struct {
	botRunner bot.Runner
}

// Option changes how the server is built.
type Option func(*options)

// WithBotRunner replaces the model runner bots use.
func WithBotRunner(r bot.Runner) Option {
	return func(o *options) { o.botRunner = r }
}

// Start builds and serves the application. Cleanup closes every client and stops the
// server. Tests may run in parallel: each one gets its own server.
func Start(t testing.TB, opts ...Option) *Server {
	t.Helper()
	o := options{botRunner: disabledRunner{}}
	for _, opt := range opts {
		opt(&o)
	}

	a, err := app.New(app.Config{
		AssetsDir:    assetsDir(),
		WebDir:       filepath.Join(t.TempDir(), "no-web"),
		ChangelogDir: filepath.Join(t.TempDir(), "no-changelog"),
		Meta:         dto.MetaResponse{Alias: "test", Name: "test", Version: "test"},
		BotRunner:    o.botRunner,
		Logger:       testutil.TestLogger(),
	})
	if err != nil {
		t.Fatalf("build server: %v", err)
	}
	s := &Server{app: a, http: httptest.NewServer(a.Handler())}
	s.URL = s.http.URL
	t.Cleanup(s.Close)
	return s
}

// MessageTypes lists every message type clients can send.
func (s *Server) MessageTypes() []dto.MessageType {
	return s.app.MessageTypes()
}

// CloseAndWait closes a client and waits until the server has finished handling the
// disconnect, so the next check sees its effect.
func (s *Server) CloseAndWait(t testing.TB, c *Client) {
	t.Helper()
	before := s.app.ConnectionCount()
	c.Close()
	deadline := time.Now().Add(Timeout())
	for s.app.ConnectionCount() >= before {
		if time.Now().After(deadline) {
			t.Fatalf("the server did not notice %s disconnecting", c.Name)
		}
		time.Sleep(time.Millisecond)
	}
	// The connection count drops before the hub runs the leave handlers; a sync on
	// any connection orders after them.
}

// Close shuts the server down the way main does, with clients possibly still connected.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.app.Close()
		s.http.Close()
	})
}

func assetsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "assets")
}

// disabledRunner fails every model call, so no test ever reaches the claude CLI.
type disabledRunner struct{}

func (disabledRunner) Run(context.Context, bot.Invocation) (bot.Result, error) {
	return bot.Result{}, errors.New("model calls are disabled in functional tests")
}

// Request sends an HTTP request with an optional JSON body and decodes a JSON response
// into out when out is not nil. It returns the status code.
func (s *Server) Request(t testing.TB, method, path string, body, out any) int {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode %s %s body: %v", method, path, err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, s.URL+path, reader)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("decode %s %s response (status %d): %v\n%s", method, path, resp.StatusCode, err, raw)
		}
	}
	return resp.StatusCode
}

// GameOptions returns the server's default game settings.
func (s *Server) GameOptions(t testing.TB) dto.GameOptionsDto {
	t.Helper()
	var options dto.GameOptionsDto
	if status := s.Request(t, http.MethodGet, "/api/v1/game-options", nil, &options); status != http.StatusOK {
		t.Fatalf("game options: status %d", status)
	}
	return options
}

// CreateGame creates a game from the server defaults, changed by edit when it is not nil,
// and returns its ID.
func (s *Server) CreateGame(t testing.TB, edit func(*dto.GameSetupDto)) string {
	t.Helper()
	settings := s.GameOptions(t).Defaults
	if edit != nil {
		edit(&settings)
	}
	var created dto.CreateGameResponse
	status := s.Request(t, http.MethodPost, "/api/v1/games", dto.CreateGameRequest{Settings: &settings}, &created)
	if status != http.StatusCreated {
		t.Fatalf("create game: status %d", status)
	}
	return created.Game.ID
}
