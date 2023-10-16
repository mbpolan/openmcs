package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var logger *slog.Logger

type Options struct {
	LogLevel string
}

// Setup prepares the logging infrastructure for the server.
func Setup(opts Options) error {
	var level slog.Level
	switch strings.ToLower(opts.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", opts.LogLevel)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger = slog.New(handler)
	return nil
}

// Debugf logs a debug message.
func Debugf(msg string, args ...any) {
	if logger == nil {
		fmt.Printf("ERROR: logger is not initialized")
		return
	}

	logger.Debug(fmt.Sprintf(msg, args...))
}

// Infof logs an informational message.
func Infof(msg string, args ...any) {
	if logger == nil {
		fmt.Printf("ERROR: logger is not initialized")
		return
	}

	logger.Info(fmt.Sprintf(msg, args...))
}

// Warnf logs a warning message.
func Warnf(msg string, args ...any) {
	if logger == nil {
		fmt.Printf("ERROR: logger is not initialized")
		return
	}

	logger.Warn(fmt.Sprintf(msg, args...))
}

// Errorf logs an error message.
func Errorf(msg string, args ...any) {
	if logger == nil {
		fmt.Printf("ERROR: logger is not initialized")
		return
	}

	logger.Error(fmt.Sprintf(msg, args...))
}
