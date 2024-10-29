// Package config provides configuration management for the server.
// It supports configuration through environment variables and command-line flags,
// offering default values and configurable options for server behavior.
package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Default configuration values.
const (
	defaultPort            = 8080
	defaultAddress         = "localhost"
	defaultStorePath       = "/tmp/metrics-db.json"
	defaultLogLevel        = "debug"
	defaultStoreInterval   = 300 // in seconds
	defaultShutdownTimeout = 30  // in seconds
)

// serverEnvs defines the server's environment variable configuration.
// It includes parameters for server address, logging, storage, database connection,
// and other settings. The struct tags specify the expected environment variables.
type serverEnvs struct {
	Address       string `env:"ADDRESS"`           // Server address in the form host:port.
	LogLevel      string `env:"LOG_LEVEL"`         // Level of logging (e.g., "debug", "info").
	StorePath     string `env:"FILE_STORAGE_PATH"` // Path to the metrics storage file.
	D3SN          string `env:"DATABASE_DSN"`      // Data Source Name for the database connection.
	Key           string `env:"KEY"`               // Secret key for data signing.
	StoreInterval int    `env:"STORE_INTERVAL"`    // Interval for storing metrics, in seconds.
	ReStore       bool   `env:"RESTORE"`           // Indicates if metrics should be restored on startup.
}

// parseServerEnvs parses server configuration from command-line flags and environment variables.
// It first sets up default values, then overrides them based on flags and environment variables.
// Returns a populated serverEnvs struct or an error if parsing fails.
func parseServerEnvs() (*serverEnvs, error) {
	envConfig := &serverEnvs{
		Address:       fmt.Sprintf("%s:%d", defaultAddress, defaultPort),
		LogLevel:      defaultLogLevel,
		StoreInterval: defaultStoreInterval,
		StorePath:     defaultStorePath,
		ReStore:       true,
	}

	flag.StringVar(&envConfig.Address, "a", envConfig.Address, "Address and port to run the server.")
	flag.StringVar(&envConfig.LogLevel, "l", envConfig.LogLevel, "Log level (e.g., debug, info, warn).")
	flag.IntVar(&envConfig.StoreInterval, "i", envConfig.StoreInterval, "Metrics store interval in seconds.")
	flag.StringVar(&envConfig.StorePath, "f", envConfig.StorePath, "Path to the metrics store file.")
	flag.BoolVar(&envConfig.ReStore, "r", envConfig.ReStore, "Enable or disable metrics restoration at startup.")
	flag.StringVar(&envConfig.D3SN, "d", "", "Database connection string (DSN).")
	flag.StringVar(&envConfig.Key, "k", "", "Secret key for signing data.")

	flag.Parse()

	if err := env.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("failed to parse server configurations: %w", err)
	}

	return envConfig, nil
}

// ServerConfig represents the full server configuration, including
// parsed environment variables and additional settings like the shutdown timeout.
type ServerConfig struct {
	Envs            *serverEnvs // Server environment configuration.
	ShutdownTimeout int         // Timeout for server shutdown, in seconds.
}

// NewServerConfig creates a new ServerConfig instance by parsing environment
// variables and command-line flags. It returns the configuration or an error
// if parsing fails.
func NewServerConfig() (*ServerConfig, error) {
	envs, err := parseServerEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to create the server configuration: %w", err)
	}

	cfg := &ServerConfig{
		Envs:            envs,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	return cfg, nil
}
