// Package logging provides a single-point constructor for the structured logger
// used by every package in the application. It exists so the composition root
// can build a *slog.Logger once and thread it through constructors, rather than
// relying on the global default logger.
package logging

import (
	"log/slog"
	"os"
)

// New returns a *slog.Logger writing text-format structured logs to stderr at
// the configured level. The composition root may replace this with a
// JSON-handler, a multi-writer, or any other slog.Handler without touching a
// single consumer package.
func New(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
