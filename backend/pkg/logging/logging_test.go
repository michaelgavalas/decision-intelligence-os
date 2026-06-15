package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/logging"
)

func TestNewWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, "info")
	log.Info("hello world", "key", "value")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %q", err, buf.String())
	}
	if entry["msg"] != "hello world" {
		t.Errorf("msg = %v, want %q", entry["msg"], "hello world")
	}
	if _, ok := entry["level"]; !ok {
		t.Error("output missing level field")
	}
}

func TestDebugSuppressedAtInfo(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, "info")
	log.Debug("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Errorf("debug message logged at info level: %q", buf.String())
	}
}

func TestDebugVisibleAtDebug(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, "debug")
	log.Debug("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("debug message missing at debug level: %q", buf.String())
	}
}

func TestLevelFromString(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"bogus", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := logging.LevelFromString(tt.in); got != tt.want {
				t.Errorf("LevelFromString(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
