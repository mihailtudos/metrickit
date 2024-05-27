package config

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"log/slog"
	"os"
)

const DefaultPort = 8080
const DefaultAddress = "localhost"
const DefaultStorePath = "/tmp/metrics-db.json"
const DefaultLogLevel = "Debug"

type serverEnvs struct {
	Address       string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StorePath     string `env:"FILE_STORAGE_PATH"`
	ReStore       bool   `env:"RESTORE"`
}

func parseServerEnvs() (*serverEnvs, error) {
	envConfig := &serverEnvs{
		Address:       DefaultAddress,
		LogLevel:      DefaultLogLevel,
		StoreInterval: 0,
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	envs, err := parseServerEnvs()
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to parse the flags", slog.String("err", err.Error()))
		return nil, errors.New("failed to create the config " + err.Error())
	}

	return &ServerConfig{
		Log:        logger,
		serverEnvs: envs,
	}, nil
}
