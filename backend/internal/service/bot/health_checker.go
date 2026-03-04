package bot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

var greetingPersonalities = map[string]string{
	"normal":  "You are friendly and casual. Use simple, enthusiastic language. Say things like 'Hey!', 'Let's go!', 'This is gonna be fun!'",
	"hard":    "You are respectful but competitive. Confident, concise. Acknowledge the challenge ahead.",
	"extreme": "You are intimidating and precise. Cold, calculated. Short, sharp statements. You exude quiet dominance.",
}

// HealthChecker verifies that a Claude API key is valid by generating a greeting.
type HealthChecker struct {
	logger *zap.Logger
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{logger: logger}
}

// CheckHealth runs a Claude CLI prompt that both verifies the API key works and generates a lobby greeting.
// Returns the greeting message on success.
func (hc *HealthChecker) CheckHealth(ctx context.Context, apiKey, model, botName, difficulty string) (string, error) {
	if model == "" {
		model = "sonnet"
	}

	personality := greetingPersonalities[difficulty]
	if personality == "" {
		personality = greetingPersonalities["normal"]
	}

	prompt := fmt.Sprintf(
		"You are %s, an AI bot joining a Terraforming Mars game lobby. %s Write a short greeting (1 sentence, max 15 words). Output ONLY the greeting, nothing else.",
		botName, personality,
	)

	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "claude",
		"-p",
		"--model", model,
		"--output-format", "text",
		prompt,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("CLAUDE_CODE_OAUTH_TOKEN=%s", apiKey))

	output, err := cmd.CombinedOutput()
	if err != nil {
		hc.logger.Error("Health check failed",
			zap.Error(err),
			zap.String("output", string(output)))
		return "", fmt.Errorf("claude health check failed: %w", err)
	}

	greeting := strings.TrimSpace(string(output))
	hc.logger.Info("Health check passed", zap.String("greeting", greeting))
	return greeting, nil
}
