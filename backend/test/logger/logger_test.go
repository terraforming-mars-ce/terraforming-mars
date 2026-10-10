package logger_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"openmars/internal/logger"
)

func TestInit_LevelFiltersMessages(t *testing.T) {
	for level, want := range map[string][]string{
		"debug":   {"debug message", "info message", "warn message", "error message"},
		"info":    {"info message", "warn message", "error message"},
		"warn":    {"warn message", "error message"},
		"error":   {"error message"},
		"unknown": {"info message", "warn message", "error message"},
	} {
		t.Run(level, func(t *testing.T) {
			var out bytes.Buffer
			logger.Init(&out, level)
			log := logger.Get()
			log.Debug("debug message")
			log.Info("info message")
			log.Warn("warn message")
			log.Error("error message")

			for _, msg := range []string{"debug message", "info message", "warn message", "error message"} {
				shown := strings.Contains(out.String(), msg)
				expected := false
				for _, w := range want {
					expected = expected || w == msg
				}
				if shown != expected {
					t.Errorf("level %s: %q shown=%v, want %v", level, msg, shown, expected)
				}
			}
		})
	}
}

func TestInit_ProductionWritesJSONLines(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	var out bytes.Buffer
	logger.Init(&out, "info")
	logger.WithGameContext("game-1", "player-1").Info("Card played")

	var line map[string]any
	if err := json.Unmarshal(out.Bytes(), &line); err != nil {
		t.Fatalf("production logs should be JSON lines: %v\n%s", err, out.String())
	}
	for key, want := range map[string]string{"msg": "Card played", "level": "INFO", "game_id": "game-1", "player_id": "player-1"} {
		if line[key] != want {
			t.Errorf("%s: got %v, want %q", key, line[key], want)
		}
	}
}

func TestContextLoggers_AddOnlyTheIDsGiven(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	cases := map[string]struct {
		log     func() // writes one line
		present []string
		absent  []string
	}{
		"game and player": {func() { logger.WithGameContext("g", "p").Info("x") }, []string{"game_id", "player_id"}, nil},
		"game only":       {func() { logger.WithGameContext("g", "").Info("x") }, []string{"game_id"}, []string{"player_id"}},
		"client":          {func() { logger.WithClientContext("c", "", "g").Info("x") }, []string{"client_id", "game_id"}, []string{"player_id"}},
		"custom":          {func() { logger.WithContext("service", "test").Info("x") }, []string{"service"}, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			logger.Init(&out, "info")
			tc.log()
			var line map[string]any
			if err := json.Unmarshal(out.Bytes(), &line); err != nil {
				t.Fatalf("decode: %v", err)
			}
			for _, key := range tc.present {
				if _, ok := line[key]; !ok {
					t.Errorf("%s should be logged", key)
				}
			}
			for _, key := range tc.absent {
				if _, ok := line[key]; ok {
					t.Errorf("an empty %s should be left out", key)
				}
			}
		})
	}
}
