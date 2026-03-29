package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/skys-mission/key-agent/internal/config"
)

func TestInit_DefaultLogger(t *testing.T) {
	// Reset the logger state
	defaultLogger = nil

	cfg := &config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		MaxSize:    10,
		MaxBackups: 1,
		MaxAge:     1,
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if defaultLogger == nil {
		t.Error("Init() did not set defaultLogger")
	}
}

func TestInit_TextFormat(t *testing.T) {
	// Reset the logger state
	defaultLogger = nil

	cfg := &config.LoggingConfig{
		Level:      "debug",
		Format:     "text",
		MaxSize:    10,
		MaxBackups: 1,
		MaxAge:     1,
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if defaultLogger == nil {
		t.Error("Init() did not set defaultLogger")
	}
}

func TestInit_FileOutput(t *testing.T) {
	// Reset the logger state
	defaultLogger = nil

	// Create temp file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := &config.LoggingConfig{
		Level:      "info",
		File:       logFile,
		Format:     "json",
		MaxSize:    10,
		MaxBackups: 1,
		MaxAge:     1,
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Log something
	Info("test message", "key", "value")

	// Check file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLogFunctions(t *testing.T) {
	// Reset and init with a buffer to capture output
	defaultLogger = nil

	var buf bytes.Buffer
	defaultLogger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Test each log function
	Debug("debug message")
	if !json.Valid(buf.Bytes()) {
		t.Error("Debug output is not valid JSON")
	}
	buf.Reset()

	Info("info message")
	if !json.Valid(buf.Bytes()) {
		t.Error("Info output is not valid JSON")
	}
	buf.Reset()

	Warn("warn message")
	if !json.Valid(buf.Bytes()) {
		t.Error("Warn output is not valid JSON")
	}
	buf.Reset()

	Error("error message")
	if !json.Valid(buf.Bytes()) {
		t.Error("Error output is not valid JSON")
	}
}

func TestWith(t *testing.T) {
	// Reset
	defaultLogger = nil

	var buf bytes.Buffer
	defaultLogger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger := With("component", "test")
	logger.Info("test message")

	output := buf.String()
	if !json.Valid([]byte(output)) {
		t.Error("Output is not valid JSON")
	}
}

func TestL(t *testing.T) {
	// L() should always return a valid logger
	logger := L()
	if logger == nil {
		t.Error("L() returned nil")
	}
}
