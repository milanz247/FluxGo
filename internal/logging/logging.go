// Package logging configures the application's structured logger.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config controls the logger's minimum level and output encoding.
type Config struct {
	// Level is one of "debug", "info", "warn", or "error". Defaults to "info".
	Level string
	// Format is "json" or "text". Defaults to "text".
	Format string
	// Output is where log records are written. Defaults to os.Stdout.
	Output io.Writer
}

// New builds a structured slog.Logger from Config. It does not modify the
// package-level default logger; call slog.SetDefault with the result to do so.
func New(config Config) *slog.Logger {
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	options := &slog.HandlerOptions{Level: parseLevel(config.Level)}

	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(config.Format), "json") {
		handler = slog.NewJSONHandler(output, options)
	} else {
		handler = slog.NewTextHandler(output, options)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
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
