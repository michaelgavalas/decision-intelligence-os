// Package logging provides a JSON structured logger built on log/slog.
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// New returns a JSON slog.Logger at the given level, writing to w. Unknown
// level strings default to info.
func New(w io.Writer, level string) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: LevelFromString(level),
	})
	return slog.New(handler)
}

// LevelFromString parses a level string, defaulting to slog.LevelInfo for
// unknown or empty input.
func LevelFromString(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
