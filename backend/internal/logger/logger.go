package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// Init initializes the global logger
func Init(logLevel *string) error {
	var err error

	var appliedLogLevel string
	if logLevel != nil {
		appliedLogLevel = *logLevel
	} else {
		appliedLogLevel = "info"
	}

	var level zapcore.Level
	switch appliedLogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	env := os.Getenv("GO_ENV")
	if env == "production" {
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(level)
		globalLogger, err = config.Build(zap.AddStacktrace(zap.ErrorLevel))
		if err != nil {
			return err
		}
	} else {
		output, _, err := zap.Open("stderr")
		if err != nil {
			return err
		}
		core := newPrettyCore(level, output)
		globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}

	return nil
}

// Get returns the global logger
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to development logger if not initialized
		globalLogger, _ = zap.NewDevelopment()
	}
	return globalLogger
}

// Sync flushes the logger
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Shutdown properly closes the logger
func Shutdown() error {
	return Sync()
}

// WithContext returns a logger with additional context fields
func WithContext(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// WithGameContext returns a logger with game-related context
func WithGameContext(gameID, playerID string) *zap.Logger {
	fields := make([]zap.Field, 0, 2)

	if gameID != "" {
		fields = append(fields, zap.String("game_id", gameID))
	}

	if playerID != "" {
		fields = append(fields, zap.String("player_id", playerID))
	}

	return Get().With(fields...)
}

// WithClientContext returns a logger with client-related context
func WithClientContext(clientID, playerID, gameID string) *zap.Logger {
	fields := make([]zap.Field, 0, 3)

	if clientID != "" {
		fields = append(fields, zap.String("client_id", clientID))
	}

	if playerID != "" {
		fields = append(fields, zap.String("player_id", playerID))
	}

	if gameID != "" {
		fields = append(fields, zap.String("game_id", gameID))
	}

	return Get().With(fields...)
}
