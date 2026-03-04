package bot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// HealthChecker verifies that a Claude API key is valid by running a quick CLI invocation.
type HealthChecker struct {
	logger *zap.Logger
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{logger: logger}
}

// CheckHealth runs a trivial Claude CLI prompt to verify the API key works.
func (hc *HealthChecker) CheckHealth(ctx context.Context, apiKey string, model string) error {
	if model == "" {
		model = "sonnet"
	}

	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "claude",
		"-p",
		"--model", model,
		"--output-format", "text",
		"respond with OK",
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("CLAUDE_CODE_OAUTH_TOKEN=%s", apiKey))

	output, err := cmd.CombinedOutput()
	if err != nil {
		hc.logger.Error("Health check failed",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("claude health check failed: %w", err)
	}

	hc.logger.Info("Health check passed", zap.String("output", string(output)))
	return nil
}
