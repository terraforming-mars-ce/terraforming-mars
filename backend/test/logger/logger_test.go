package logger_test

import (
	"testing"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
)

func TestInit(t *testing.T) {
	// Test development environment
	t.Setenv("GO_ENV", "development")
	err := logger.Init(nil)
	if err != nil {
		t.Fatalf("Failed to initialize logger in development mode: %v", err)
	}

	log := logger.Get()
	if log == nil {
		t.Fatal("Logger should not be nil after initialization")
	}

	// Test production environment
	t.Setenv("GO_ENV", "production")
	err = logger.Init(nil)
	if err != nil {
		t.Fatalf("Failed to initialize logger in production mode: %v", err)
	}

	log = logger.Get()
	if log == nil {
		t.Fatal("Logger should not be nil after initialization")
	}

	// Clean up
	if err := logger.Shutdown(); err != nil {
		t.Logf("logger shutdown: %v", err)
	}
}

func TestWithGameContext(t *testing.T) {
	err := logger.Init(nil)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() { _ = logger.Shutdown() }()

	gameID := "test-game-123"
	playerID := "test-player-456"

	contextLogger := logger.WithGameContext(gameID, playerID)
	if contextLogger == nil {
		t.Fatal("Context logger should not be nil")
	}

	// Test with empty values
	contextLogger = logger.WithGameContext("", "")
	if contextLogger == nil {
		t.Fatal("Context logger should not be nil even with empty values")
	}
}

func TestWithClientContext(t *testing.T) {
	err := logger.Init(nil)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() { _ = logger.Shutdown() }()

	clientID := "client-123"
	playerID := "player-456"
	gameID := "game-789"

	contextLogger := logger.WithClientContext(clientID, playerID, gameID)
	if contextLogger == nil {
		t.Fatal("Context logger should not be nil")
	}

	// Test that we can log with the context logger without panic
	contextLogger.Info("Test message with client context",
		zap.String("test_field", "test_value"),
	)
}

func TestLoggerFallback(t *testing.T) {
	// Test that Get() returns a logger even if Init() wasn't called
	log := logger.Get()
	if log == nil {
		t.Fatal("Get() should return a fallback logger if not initialized")
	}

	// Verify we can log without panic
	log.Info("Fallback logger test")
}

func TestWithContext(t *testing.T) {
	err := logger.Init(nil)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() { _ = logger.Shutdown() }()

	contextLogger := logger.WithContext(
		zap.String("service", "test"),
		zap.Int("version", 1),
	)

	if contextLogger == nil {
		t.Fatal("Context logger should not be nil")
	}

	// Test that we can log with the context logger
	contextLogger.Info("Test message with custom context")
}
