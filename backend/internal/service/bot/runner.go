package bot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// MCPServerName is the name the bot's game tool server is registered under in the CLI.
// Tools appear to the model as mcp__<MCPServerName>__<tool>.
const MCPServerName = "openmars"

// MCPEndpoint is the per-invocation address and credential of the game tool server.
type MCPEndpoint struct {
	URL    string
	Bearer string
}

// Invocation describes one model call.
type Invocation struct {
	Model        string
	SystemPrompt string
	Prompt       string
	Token        string
	// MCP is nil for text-only calls (reactions, greetings, recaps).
	MCP          *MCPEndpoint
	MaxBudgetUSD float64
	// OnText receives each assistant text block as it streams.
	OnText func(text string)
}

// Result is the outcome of one model call.
type Result struct {
	Text     string
	CostUSD  float64
	Turns    int
	Duration time.Duration
}

// Runner runs model calls. The CLI implementation is replaced by fakes in tests.
type Runner interface {
	Run(ctx context.Context, inv Invocation) (Result, error)
}

// CLIRunner runs model calls through the claude CLI with no built-in tools,
// only the game tool server, an isolated working directory and a minimal environment.
type CLIRunner struct {
	binary string
	logger *slog.Logger
}

// NewCLIRunner creates a runner that invokes the claude binary found on PATH.
func NewCLIRunner(logger *slog.Logger) *CLIRunner {
	return &CLIRunner{binary: "claude", logger: logger}
}

// BuildCLIArgs returns the CLI arguments for an invocation. The prompt is sent on stdin
// because several flags are variadic and would swallow a trailing positional argument.
func BuildCLIArgs(inv Invocation) ([]string, error) {
	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--model", inv.Model,
		"--tools", "",
		"--strict-mcp-config",
		"--no-session-persistence",
		"--permission-mode", "dontAsk",
		"--system-prompt", inv.SystemPrompt,
	}
	if inv.MaxBudgetUSD > 0 {
		args = append(args, "--max-budget-usd", strconv.FormatFloat(inv.MaxBudgetUSD, 'f', 4, 64))
	}
	if inv.MCP != nil {
		config, err := json.Marshal(map[string]any{
			"mcpServers": map[string]any{
				MCPServerName: map[string]any{
					"type":    "http",
					"url":     inv.MCP.URL,
					"headers": map[string]string{"Authorization": "Bearer " + inv.MCP.Bearer},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("encode mcp config: %w", err)
		}
		args = append(args, "--mcp-config", string(config), "--allowedTools", "mcp__"+MCPServerName+"__*")
	}
	return args, nil
}

// BuildCLIEnv returns the complete environment for the CLI process. Nothing is
// inherited from the server process except PATH, so no server secret reaches the model.
func BuildCLIEnv(token, home, path string) []string {
	return []string{
		"PATH=" + path,
		"HOME=" + home,
		"CLAUDE_CODE_OAUTH_TOKEN=" + token,
		"DISABLE_AUTOUPDATER=1",
		"DISABLE_TELEMETRY=1",
	}
}

// ErrAuthentication means the Claude token was rejected.
var ErrAuthentication = errors.New("the Claude token was rejected")

type streamEvent struct {
	Type        string         `json:"type"`
	ErrorStatus int            `json:"error_status,omitempty"`
	Subtype     string         `json:"subtype,omitempty"`
	Message     *streamMessage `json:"message,omitempty"`
	Result      string         `json:"result,omitempty"`
	IsError     bool           `json:"is_error,omitempty"`
	Cost        float64        `json:"total_cost_usd,omitempty"`
	NumTurns    int            `json:"num_turns,omitempty"`
	DurationMs  int            `json:"duration_ms,omitempty"`
}

type streamMessage struct {
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// Run executes the CLI and returns its final text and cost.
func (r *CLIRunner) Run(ctx context.Context, inv Invocation) (Result, error) {
	args, err := BuildCLIArgs(inv)
	if err != nil {
		return Result{}, err
	}

	workDir, err := os.MkdirTemp("", "openmars-bot-")
	if err != nil {
		return Result{}, fmt.Errorf("create bot work dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			r.logger.Warn("Failed to remove bot work dir", slog.String("path", workDir), slog.Any("error", err))
		}
	}()

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()
	cmd := exec.CommandContext(runCtx, r.binary, args...)
	cmd.Dir = workDir
	cmd.Env = BuildCLIEnv(inv.Token, workDir, os.Getenv("PATH"))
	cmd.Stdin = strings.NewReader(inv.Prompt)
	var stderr bytes.Buffer
	cmd.Stderr = &limitedWriter{buf: &stderr, limit: 4096}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start claude CLI: %w", err)
	}

	var result Result
	var resultErr error
	gotResult := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 4*1024*1024)
	for scanner.Scan() {
		var event streamEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		switch event.Type {
		case "system":
			if event.Subtype == "api_retry" && (event.ErrorStatus == 401 || event.ErrorStatus == 403) {
				resultErr = ErrAuthentication
				cancelRun()
			}
		case "assistant":
			r.handleAssistant(event, inv.OnText)
		case "result":
			gotResult = true
			result = Result{
				Text:     event.Result,
				CostUSD:  event.Cost,
				Turns:    event.NumTurns,
				Duration: time.Duration(event.DurationMs) * time.Millisecond,
			}
			if event.IsError || event.Subtype != "success" {
				resultErr = fmt.Errorf("claude %s: %s", event.Subtype, truncate(event.Result, 300))
			}
		}
	}

	waitErr := cmd.Wait()
	if errors.Is(resultErr, ErrAuthentication) {
		return result, resultErr
	}
	if ctx.Err() != nil {
		return result, fmt.Errorf("claude CLI cancelled: %w", ctx.Err())
	}
	if resultErr != nil {
		return result, resultErr
	}
	if !gotResult {
		return result, fmt.Errorf("claude CLI ended without a result: %v: %s", waitErr, truncate(stderr.String(), 300))
	}
	r.logger.Debug("Model call finished",
		slog.String("model", inv.Model),
		slog.Int("turns", result.Turns),
		slog.Duration("duration", result.Duration),
		slog.Float64("cost_usd", result.CostUSD))
	return result, nil
}

func (r *CLIRunner) handleAssistant(event streamEvent, onText func(string)) {
	if event.Message == nil {
		return
	}
	var blocks []contentBlock
	if err := json.Unmarshal(event.Message.Content, &blocks); err != nil {
		return
	}
	for _, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text != "" && onText != nil {
				onText(block.Text)
			}
		case "tool_use":
			r.logger.Debug("Bot tool call", slog.String("tool", block.Name), slog.String("input", truncate(string(block.Input), 300)))
		}
	}
}

type limitedWriter struct {
	buf   *bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		if len(p) > room {
			w.buf.Write(p[:room])
		} else {
			w.buf.Write(p)
		}
	}
	return len(p), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
