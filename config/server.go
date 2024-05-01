package config

import (
	"log/slog"
	"os"
)

type AppConfig struct {
	Address string
	Log     *slog.Logger
}

func NewAppConfig(port string) AppConfig {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return AppConfig{Address: port, Log: logger}
}
