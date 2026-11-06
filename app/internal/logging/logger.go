// Package logging builds the structured JSON logger used by the service.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures the JSON logger.
type Options struct {
	Service string
	Version string
	Commit  string
	Level   string
	Output  io.Writer
}

// New returns a JSON logger enriched with the service build metadata.
func New(opts Options) *slog.Logger {
	output := opts.Output
	if output == nil {
		output = os.Stdout
	}

	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: parseLevel(opts.Level),
	})

	return slog.New(handler).With(
		slog.String("service", opts.Service),
		slog.String("version", opts.Version),
		slog.String("commit", opts.Commit),
	)
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
