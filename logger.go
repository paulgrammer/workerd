package workerd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// LoggerFactory handles logger creation and configuration
type LoggerFactory struct{}

// NewLoggerFactory creates a new logger factory
func NewLoggerFactory() *LoggerFactory {
	return &LoggerFactory{}
}

// CreateLogger creates a new structured logger with the specified level
func (lf *LoggerFactory) CreateLogger(level slog.Level) *slog.Logger {
	handler := &structuredLogHandler{level: level}
	return slog.New(handler)
}

// CreateDefaultLogger creates a logger with INFO level
func (lf *LoggerFactory) CreateDefaultLogger() *slog.Logger {
	return lf.CreateLogger(slog.LevelInfo)
}

// structuredLogHandler implements slog.Handler for custom log formatting
type structuredLogHandler struct {
	level slog.Level
}

// Enabled reports whether the handler handles records at the given level
func (h *structuredLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the log record
func (h *structuredLogHandler) Handle(_ context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006/01/02 15:04:05.000000")
	pid := os.Getpid()
	level := h.formatLevel(r.Level)

	// Build attributes string
	attrs := h.buildAttributes(r)
	
	if attrs != "" {
		fmt.Printf("pid=%d %s %s %s %s\n", pid, timestamp, level, r.Message, attrs)
	} else {
		fmt.Printf("pid=%d %s %s %s\n", pid, timestamp, level, r.Message)
	}
	
	return nil
}

// formatLevel formats the log level for display
func (h *structuredLogHandler) formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelInfo:
		return "INFO:"
	case slog.LevelError:
		return "ERROR:"
	case slog.LevelWarn:
		return "WARN:"
	case slog.LevelDebug:
		return "DEBUG:"
	default:
		return level.String() + ":"
	}
}

// buildAttributes builds a string representation of log attributes
func (h *structuredLogHandler) buildAttributes(r slog.Record) string {
	if r.NumAttrs() == 0 {
		return ""
	}

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	result := ""
	for i, attr := range attrs {
		if i > 0 {
			result += " "
		}
		result += attr
	}
	
	return result
}

// WithAttrs returns a new handler with additional attributes
func (h *structuredLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, returning the same handler
	// In a more complex implementation, you might want to store and use these attrs
	return h
}

// WithGroup returns a new handler with a group name
func (h *structuredLogHandler) WithGroup(name string) slog.Handler {
	// For simplicity, returning the same handler
	// In a more complex implementation, you might want to handle groups
	return h
}

// Global logger factory instance
var defaultLoggerFactory = NewLoggerFactory()

// newLogger creates a new logger using the global factory
func newLogger(level slog.Level) *slog.Logger {
	return defaultLoggerFactory.CreateLogger(level)
}