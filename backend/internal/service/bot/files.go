package bot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HistoryWriter appends messages to a log file for Claude to verify command results.
type HistoryWriter struct {
	file   *os.File
	mu     sync.Mutex
	logger *zap.Logger
}

// NewHistoryWriter creates a new history writer that truncates the log file.
func NewHistoryWriter(path string, logger *zap.Logger) (*HistoryWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open history file: %w", err)
	}
	return &HistoryWriter{file: f, logger: logger}, nil
}

// WriteReceived logs an inbound message (server response).
func (h *HistoryWriter) WriteReceived(msgType string, data json.RawMessage) {
	h.write("<<", msgType, data)
}

// WriteSent logs an outbound message (bot command).
func (h *HistoryWriter) WriteSent(msgType string, data json.RawMessage) {
	h.write(">>", msgType, data)
}

func (h *HistoryWriter) write(direction, msgType string, data json.RawMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ts := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s %s %s\n", ts, direction, msgType, string(data))
	if _, err := h.file.WriteString(line); err != nil {
		h.logger.Error("Failed to write history", zap.Error(err))
	}
}

// Close closes the history file.
func (h *HistoryWriter) Close() error {
	return h.file.Close()
}

// StateWriter atomically writes the game state summary to a file.
type StateWriter struct {
	path string
}

// NewStateWriter creates a new state writer.
func NewStateWriter(path string) *StateWriter {
	return &StateWriter{path: path}
}

// WriteState atomically writes the summary to the state file.
func (w *StateWriter) WriteState(summary string) error {
	tmpPath := w.path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("write tmp state: %w", err)
	}
	if err := os.Rename(tmpPath, w.path); err != nil {
		return fmt.Errorf("rename state file: %w", err)
	}
	return nil
}

// CommandReader polls a JSONL file for new commands written by Claude.
type CommandReader struct {
	path     string
	commands chan json.RawMessage
	done     chan struct{}
	offset   int64
	logger   *zap.Logger
}

// NewCommandReader creates a new command reader.
func NewCommandReader(path string, logger *zap.Logger) *CommandReader {
	return &CommandReader{
		path:     path,
		commands: make(chan json.RawMessage, 32),
		done:     make(chan struct{}),
		logger:   logger,
	}
}

// Start creates the command file and begins polling.
func (r *CommandReader) Start() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}
	f, err := os.Create(r.path)
	if err != nil {
		return fmt.Errorf("create command file: %w", err)
	}
	if err := f.Close(); err != nil {
		r.logger.Warn("Failed to close command file after creation", zap.Error(err))
	}

	go r.watch()
	return nil
}

// Commands returns a channel of raw JSON commands.
func (r *CommandReader) Commands() <-chan json.RawMessage {
	return r.commands
}

// Stop stops the command reader.
func (r *CommandReader) Stop() {
	close(r.done)
}

// Reset truncates the command file and resets the offset.
func (r *CommandReader) Reset() error {
	f, err := os.Create(r.path)
	if err != nil {
		return fmt.Errorf("reset command file: %w", err)
	}
	if err := f.Close(); err != nil {
		r.logger.Warn("Failed to close command file after reset", zap.Error(err))
	}
	r.offset = 0
	return nil
}

func (r *CommandReader) watch() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.readNewLines()
		}
	}
}

func (r *CommandReader) readNewLines() {
	f, err := os.Open(r.path)
	if err != nil {
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			r.logger.Warn("Failed to close command file after reading", zap.Error(err))
		}
	}()

	if _, err := f.Seek(r.offset, io.SeekStart); err != nil {
		return
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw json.RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			r.logger.Warn("Invalid JSON line in command file", zap.String("line", line))
			continue
		}

		select {
		case r.commands <- raw:
		case <-r.done:
			return
		}
	}

	newOffset, _ := f.Seek(0, io.SeekCurrent)
	r.offset = newOffset
}
