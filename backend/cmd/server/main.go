package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"openmars/internal/app"
	"openmars/internal/delivery/dto"
	"openmars/internal/logger"

	"github.com/joho/godotenv"
)

var Version = "localbuild"

var serverAliasPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// loadServerMeta reads the identity a gateway uses to list and verify this server.
func loadServerMeta() (dto.MetaResponse, error) {
	alias := os.Getenv("OPENMARS_SERVER_ALIAS")
	if alias == "" {
		alias = "local"
	}
	if !serverAliasPattern.MatchString(alias) {
		return dto.MetaResponse{}, fmt.Errorf("OPENMARS_SERVER_ALIAS %q must match %s", alias, serverAliasPattern)
	}
	name := os.Getenv("OPENMARS_SERVER_NAME")
	if name == "" {
		name = alias
	}
	return dto.MetaResponse{Alias: alias, Name: name, Version: Version}, nil
}

func main() {
	// Load env from dev.env if present (does not override existing env vars)
	_ = godotenv.Load("dev.env")

	logLevel := os.Getenv("OPENMARS_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logger.Init(os.Stderr, logLevel)

	log := logger.Get()
	log.Info("Starting Open Mars backend server")
	log.Info("Version: " + Version)
	log.Debug("Log level set to " + logLevel)

	addr := os.Getenv("OPENMARS_ADDR")
	if addr == "" {
		addr = ":3001"
	}
	webDir := os.Getenv("OPENMARS_WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}

	serverMeta, err := loadServerMeta()
	if err != nil {
		log.Error("Invalid server identity", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Server alias: " + serverMeta.Alias)

	changelogDir := os.Getenv("OPENMARS_CHANGELOG_DIR")
	if changelogDir == "" {
		changelogDir = "changelog"
	}
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wd, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get working directory", slog.Any("error", err))
		os.Exit(1)
	}

	application, err := app.New(app.Config{
		AssetsDir:    filepath.Join(wd, "assets"),
		WebDir:       webDir,
		ChangelogDir: changelogDir,
		Meta:         serverMeta,
		Logger:       log,
	})
	if err != nil {
		log.Error("Failed to build server", slog.Any("error", err))
		os.Exit(1)
	}

	// ========== Setup HTTP Server ==========
	server := &http.Server{
		Addr:         addr,
		Handler:      application.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Info("HTTP server listening", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start HTTP server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	log.Info("Server started")

	// Wait for shutdown signal
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to gracefully shutdown HTTP server", slog.Any("error", err))
	} else {
		log.Debug("HTTP server stopped")
	}

	application.Close()
	log.Debug("WebSocket hub stopped")

	log.Info("Server shutdown complete")
}
