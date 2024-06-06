package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func NewLogger(w io.Writer, level string) *slog.Logger {
	l := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if level != "" {
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: getLevel(level, l)}))
	}

	l.DebugContext(context.Background(),
		"logger created with default log level",
		slog.String("level", "debug"),
	)
	return l
}

func getLevel(level string, logger *slog.Logger) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		logger.ErrorContext(context.Background(),
			fmt.Sprintf("log level '%s' not found, defaulting to '%s'", level, defaultLogLevel),
		)
		return getLevel(defaultLogLevel, logger)
	}
}
