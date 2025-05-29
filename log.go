package workerd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type slogHandler struct {
	level slog.Level
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006/01/02 15:04:05.000000")
	pid := os.Getpid()
	level := r.Level.String()

	msg := r.Message
	if level == "INFO" {
		level = "INFO:"
	} else if level == "ERROR" {
		level = "ERROR:"
	} else {
		level = level + ":"
	}

	fmt.Printf("pid=%d %s %s %s\n", pid, timestamp, level, msg)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // ignoring attrs for simplicity
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return h // ignoring groups for simplicity
}

func newLogger(level slog.Level) *slog.Logger {
	return slog.New(&slogHandler{level: level})
}
