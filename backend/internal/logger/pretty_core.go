package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

// ANSI color codes
const (
	ansiReset     = "\033[0m"
	ansiDim       = "\033[2m"
	ansiGrey      = "\033[90m"
	ansiLightGrey = "\033[37m"
	ansiCyan      = "\033[36m"
	ansiBlue      = "\033[34m"
	ansiYellow    = "\033[33m"
	ansiRed       = "\033[31m"
)

// prettyCore is a custom zapcore.Core that formats log output with colors
// and places structured fields on a separate line aligned with the caller.
type prettyCore struct {
	zapcore.LevelEnabler
	output zapcore.WriteSyncer
	fields []zapcore.Field
}

func newPrettyCore(level zapcore.LevelEnabler, output zapcore.WriteSyncer) *prettyCore {
	return &prettyCore{
		LevelEnabler: level,
		output:       output,
	}
}

func (c *prettyCore) With(fields []zapcore.Field) zapcore.Core {
	clone := &prettyCore{
		LevelEnabler: c.LevelEnabler,
		output:       c.output,
		fields:       make([]zapcore.Field, len(c.fields)+len(fields)),
	}
	copy(clone.fields, c.fields)
	copy(clone.fields[len(c.fields):], fields)
	return clone
}

func (c *prettyCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *prettyCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	allFields := append(c.fields[:len(c.fields):len(c.fields)], fields...)

	var sb strings.Builder

	// Timestamp (dim)
	ts := entry.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	sb.WriteString(ansiDim)
	sb.WriteString(ts)
	sb.WriteString(ansiReset)
	sb.WriteString("    ")

	// Level (colored, padded to 5 chars)
	levelStr := entry.Level.CapitalString()
	for len(levelStr) < 5 {
		levelStr += " "
	}
	sb.WriteString(levelColor(entry.Level))
	sb.WriteString(levelStr)
	sb.WriteString(ansiReset)
	sb.WriteString("    ")

	// Caller (dim)
	callerStr := ""
	if entry.Caller.Defined {
		callerStr = entry.Caller.TrimmedPath()
		sb.WriteString(ansiDim)
		sb.WriteString(callerStr)
		sb.WriteString(ansiReset)
		sb.WriteString("    ")
	}

	// Message (normal)
	sb.WriteString(entry.Message)

	// Structured fields as key=value pairs on same line, in grey
	if len(allFields) > 0 {
		kvStr := fieldsToKV(allFields)
		if kvStr != "" {
			sb.WriteString("    ")
			sb.WriteString(kvStr)
			sb.WriteString(ansiReset)
		}
	}

	// Stack trace
	if entry.Stack != "" {
		sb.WriteString("\n")
		sb.WriteString(ansiDim)
		sb.WriteString(entry.Stack)
		sb.WriteString(ansiReset)
	}

	sb.WriteString("\n")

	_, err := c.output.Write([]byte(sb.String()))
	return err
}

func (c *prettyCore) Sync() error {
	return c.output.Sync()
}

func levelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return ansiCyan
	case zapcore.InfoLevel:
		return ansiBlue
	case zapcore.WarnLevel:
		return ansiYellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ansiRed
	default:
		return ansiReset
	}
}

func fieldsToKV(fields []zapcore.Field) string {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}
	if len(enc.Fields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(enc.Fields))
	for k, v := range enc.Fields {
		parts = append(parts, fmt.Sprintf("%s%s=%s%v", ansiGrey, k, ansiLightGrey, v))
	}
	return strings.Join(parts, " ")
}
