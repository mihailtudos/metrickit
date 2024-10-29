// Package logger provides utilities for creating and configuring a structured logger
// using the slog library. It allows setting different log levels and output formats.
package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// NewLogger creates a new slog.Logger instance with the specified output writer
// and log level. It returns an error if the log level is invalid.
func NewLogger(w io.Writer, level string) (*slog.Logger, error) {
	ll, err := getLevel(level)
	if err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}

	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: ll})), nil
}

// getLevel parses the provided log level string and returns the corresponding slog.Level.
// It returns an error if the log level is invalid.
func getLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("incorrect log level provided")
	}
}
