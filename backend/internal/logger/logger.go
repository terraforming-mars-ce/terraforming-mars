package logger

import (
	"io"
	"log/slog"
	"os"
)

// globalLevel holds the active log level; adjusting it reconfigures the default
// logger without rebuilding the handler.
var globalLevel = new(slog.LevelVar)

// Init sets the process-wide slog default logger, writing to w at the given level
// ("debug", "info", "warn" or "error"; anything else means info). In production it emits
// JSON; otherwise it uses the pretty colored console handler. The whole backend
// — domain, action, delivery, service — logs through slog.Default().
func Init(w io.Writer, level string) {
	switch level {
	case "debug":
		globalLevel.Set(slog.LevelDebug)
	case "warn":
		globalLevel.Set(slog.LevelWarn)
	case "error":
		globalLevel.Set(slog.LevelError)
	default:
		globalLevel.Set(slog.LevelInfo)
	}

	var handler slog.Handler
	if os.Getenv("GO_ENV") == "production" {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: globalLevel, AddSource: true})
	} else {
		handler = newPrettyHandler(w, globalLevel)
	}

	slog.SetDefault(slog.New(handler))
}

// Get returns the process-wide logger. Before Init runs it falls back to the
// slog default (a plain text handler), so callers never get a nil logger.
func Get() *slog.Logger {
	return slog.Default()
}

// WithContext returns a logger with additional context attributes.
func WithContext(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

// WithGameContext returns a logger with game-related context.
func WithGameContext(gameID, playerID string) *slog.Logger {
	args := make([]any, 0, 2)
	if gameID != "" {
		args = append(args, slog.String("game_id", gameID))
	}
	if playerID != "" {
		args = append(args, slog.String("player_id", playerID))
	}
	return slog.Default().With(args...)
}

// WithClientContext returns a logger with client-related context.
func WithClientContext(clientID, playerID, gameID string) *slog.Logger {
	args := make([]any, 0, 3)
	if clientID != "" {
		args = append(args, slog.String("client_id", clientID))
	}
	if playerID != "" {
		args = append(args, slog.String("player_id", playerID))
	}
	if gameID != "" {
		args = append(args, slog.String("game_id", gameID))
	}
	return slog.Default().With(args...)
}
