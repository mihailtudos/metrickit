package config

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

const defaultPort = 8080
const defaultAddress = "localhost"
const defaultStorePath = "/tmp/metrics-db.json"
const defaultLogLevel = "debug"
const defaultStoreInterval = 300
const defaultShutdownTimeout = 30

type serverEnvs struct {
	Address       string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	StorePath     string `env:"FILE_STORAGE_PATH"`
	D3SN          string `env:"DATABASE_DSN"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	ReStore       bool   `env:"RESTORE"`
}

func parseServerEnvs() (*serverEnvs, error) {
	envConfig := &serverEnvs{
		Address:       fmt.Sprintf("%s:%d", defaultAddress, defaultPort),
		LogLevel:      defaultLogLevel,
		StoreInterval: defaultStoreInterval,
		StorePath:     defaultStorePath,
		ReStore:       true,
	}

	flag.StringVar(&envConfig.Address, "a", envConfig.Address, "address and port to run the server")
	flag.StringVar(&envConfig.LogLevel, "l", envConfig.LogLevel, "log level")
	flag.IntVar(&envConfig.StoreInterval, "i", envConfig.StoreInterval, "metrics store interval in seconds")
	flag.StringVar(&envConfig.StorePath, "f", envConfig.StorePath, "metrics store file path")
	flag.BoolVar(&envConfig.ReStore, "r", envConfig.ReStore, "metrics re-store option")
	flag.StringVar(&envConfig.D3SN, "d", "", "DB connection string")

	flag.Parse()

	if err := env.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("server configs: %w", err)
	}

	return envConfig, nil
}

type ServerConfig struct {
	Log             *slog.Logger
	Envs            *serverEnvs
	DB              *sql.DB
	ShutdownTimeout int
}

func NewServerConfig() (*ServerConfig, error) {
	logger := NewLogger(os.Stdout, defaultLogLevel)

	envs, err := parseServerEnvs()
	if err != nil {
		return nil, errors.New("failed to create the config " + err.Error())
	}

	if envs.LogLevel != defaultLogLevel {
		logger = NewLogger(os.Stdout, envs.LogLevel)
	}

	cfg := &ServerConfig{
		Log:             logger,
		Envs:            envs,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if cfg.Envs.D3SN != "" {
		db, err := cfg.InitPostgresDB(cfg.Envs.D3SN)
		if err != nil {
			return nil, fmt.Errorf("DB DSN present but failed to initiatie connection: %w", err)
		}
		cfg.DB = db
	}

	return cfg, nil
}
