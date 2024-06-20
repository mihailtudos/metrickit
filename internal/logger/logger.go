package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func NewLogger(w io.Writer, level string) (*slog.Logger, error) {
	ll, err := getLevel(level)
	if err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}

	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: ll})), nil
}

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
