// Package logger provides structured logging with rotation support.
package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/skys-mission/key-agent/internal/config"
)

var (
	// Default logger instance
	defaultLogger *slog.Logger
	once          sync.Once
)

// Init initializes the default logger with the given configuration.
func Init(cfg *config.LoggingConfig) error {
	var writer io.Writer

	if cfg.File != "" {
		// Ensure log directory exists
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		writer = &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	} else {
		writer = os.Stderr
	}

	// Parse log level
	level := parseLevel(cfg.Level)

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if cfg.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

// initDefault ensures a default logger exists.
func initDefault() {
	once.Do(func() {
		if defaultLogger == nil {
			defaultLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
			slog.SetDefault(defaultLogger)
		}
	})
}

// parseLevel converts string level to slog.Level.
func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	initDefault()
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	initDefault()
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	initDefault()
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	initDefault()
	defaultLogger.Error(msg, args...)
}

// With returns a logger with additional context.
func With(args ...any) *slog.Logger {
	initDefault()
	return defaultLogger.With(args...)
}

// L returns the default logger.
func L() *slog.Logger {
	initDefault()
	return defaultLogger
}
