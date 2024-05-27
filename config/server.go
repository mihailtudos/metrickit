package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"log/slog"
	"os"
	"strings"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"
const DefaultStorePath = "/tmp/metrics-db.json"
const DefaultLogLevel = "debug"
const DefaultStoreInterval = 300

type serverEnvs struct {
	Address       string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StorePath     string `env:"FILE_STORAGE_PATH"`
	ReStore       bool   `env:"RESTORE"`
}

func parseServerEnvs() (*serverEnvs, error) {
	envConfig := &serverEnvs{
		Address:       fmt.Sprintf("%s:%d", DefaultAddress, DefaultPort),
		LogLevel:      DefaultLogLevel,
		StoreInterval: DefaultStoreInterval,
		StorePath:     DefaultStorePath,
		ReStore:       true,
	}

	flag.StringVar(&envConfig.Address, "a", envConfig.Address, "address and port to run the server")
	flag.StringVar(&envConfig.LogLevel, "l", envConfig.LogLevel, "log level")
	flag.IntVar(&envConfig.StoreInterval, "i", envConfig.StoreInterval, "metrics store interval")
	flag.StringVar(&envConfig.StorePath, "f", envConfig.StorePath, "metrics store file path")
	flag.BoolVar(&envConfig.ReStore, "r", envConfig.ReStore, "metrics re-store option")

	flag.Parse()

	if err := env.Parse(envConfig); err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}

	return envConfig, nil
}

type ServerConfig struct {
	Log *slog.Logger
	*serverEnvs
}

func NewServerConfig() (*ServerConfig, error) {
	envs, err := parseServerEnvs()
	if err != nil {
		return nil, errors.New("failed to create the config " + err.Error())
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: getLevel(envs.LogLevel)}))

	return &ServerConfig{
		Log:        logger,
		serverEnvs: envs,
	}, nil
}

func getLevel(level string) slog.Level {
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
		return getLevel(DefaultLogLevel)
	}
}
