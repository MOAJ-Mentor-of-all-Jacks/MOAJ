package observability

import (
	"io"
	"log/slog"
)

// NewLogger creates the JSON logger shared by startup, requests, and shutdown.
// Callers log named safe fields, never whole requests, config, or panic values.
func NewLogger(out io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level}))
}
