package workerd

import (
	"log/slog"
	"os"
)

// newLogger creates a new logger using the global factory
func newLogger(level slog.Level) *slog.Logger {
	baseAttrs := []slog.Attr{slog.Int("pid", os.Getpid())}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	handlerWithPID := handler.WithAttrs(baseAttrs)
	logger := slog.New(handlerWithPID)

	return logger
}
