// Package config provides configuration management for the server.
// It supports configuration through environment variables and command-line flags,
// offering default values and configurable options for server behavior.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/internal/utils"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// Default configuration values.
const (
	defaultPort            = 8080
	defaultAddress         = "localhost"
	DefaultStorePath       = "/tmp/metrics-db.json"
	DefaultLogLevel        = "debug"
	DefaultStoreInterval   = 300 // in seconds
	defaultShutdownTimeout = 30  // in seconds
)

// serverEnvs defines the server's environment variable configuration.
// It includes parameters for server address, logging, storage, database connection,
// and other settings. The struct tags specify the expected environment variables.
type serverEnvs struct {
	Address        string `env:"ADDRESS" json:"address"`               // Server address in the form host:port.
	LogLevel       string `env:"LOG_LEVEL"`                            // Level of logging (e.g., "debug", "info").
	StorePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`  // Path to the metrics storage file.
	D3SN           string `env:"DATABASE_DSN" json:"database_dsn"`     // Data Source Name for the database connection.
	Key            string `env:"KEY"`                                  // Secret key for data signing.
	PrivateKeyPath string `env:"CRYPTO_KEY" json:"crypto_key"`         // Path to the private key file.
	ConfigPath     string `env:"CONFIG"`                               // Path to the configuration file.
	StoreInterval  int    `env:"STORE_INTERVAL" json:"store_interval"` // Interval for storing metrics, in seconds.
	// Indicates if metrics should be restored on startup.
	ReStore bool `env:"RESTORE" json:"restore"`
}

// parseServerEnvs parses server configuration from command-line flags and environment variables.
// It first sets up default values, then overrides them based on flags and environment variables.
// Returns a populated serverEnvs struct or an error if parsing fails.
func parseServerEnvs() (*serverEnvs, error) {
	envConfig := &serverEnvs{
		Address:       fmt.Sprintf("%s:%d", defaultAddress, defaultPort),
		LogLevel:      DefaultLogLevel,
		StoreInterval: DefaultStoreInterval,
		StorePath:     DefaultStorePath,
		ReStore:       true,
	}

	flag.StringVar(&envConfig.ConfigPath, "config", "", "Path to the json configuration file.")
	flag.StringVar(&envConfig.Address, "a", envConfig.Address, "Address and port to run the server.")
	flag.StringVar(&envConfig.Address, "a", envConfig.Address, "Address and port to run the server.")
	flag.StringVar(&envConfig.LogLevel, "l", envConfig.LogLevel, "Log level (e.g., debug, info, warn).")
	flag.IntVar(&envConfig.StoreInterval, "i", envConfig.StoreInterval, "Metrics store interval in seconds.")
	flag.StringVar(&envConfig.StorePath, "f", envConfig.StorePath, "Path to the metrics store file.")
	flag.BoolVar(&envConfig.ReStore, "r", envConfig.ReStore, "Enable or disable metrics restoration at startup.")
	flag.StringVar(&envConfig.D3SN, "d", "", "Database connection string (DSN).")
	flag.StringVar(&envConfig.Key, "k", "", "Secret key for signing data.")
	flag.StringVar(&envConfig.PrivateKeyPath, "crypto-key", envConfig.PrivateKeyPath, "Path to the private key file.")

	flag.Parse()

	if err := env.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("failed to parse server configurations: %w", err)
	}

	if envConfig.ConfigPath != "" {
		viper.SetConfigName("server")
		viper.SetConfigType("json")
		viper.AddConfigPath(envConfig.ConfigPath)

		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read json config file: %w", err)
		}

		utils.Replace(&envConfig.Address, viper.GetString("address"))
		utils.Replace(&envConfig.StorePath, viper.GetString("store_file"))
		utils.Replace(&envConfig.D3SN, viper.GetString("database_dsn"))
		utils.Replace(&envConfig.PrivateKeyPath, viper.GetString("crypto_key"))

		utils.Replace(&envConfig.ReStore, viper.GetBool("restore"))

		utils.Replace(&envConfig.StoreInterval, int(viper.GetDuration("store_interval").Seconds()))
	}

	return envConfig, nil
}

// ServerConfig represents the full server configuration, including
// parsed environment variables and additional settings like the shutdown timeout.
type ServerConfig struct {
	Envs            *serverEnvs     // Server environment configuration.
	PrivateKey      *rsa.PrivateKey // Private key for encryption, configurable via environment variable "CRYPTO_KEY".
	ShutdownTimeout int             // Timeout for server shutdown, in seconds.
}

// NewServerConfig creates a new ServerConfig instance by parsing environment
// variables and command-line flags. It returns the configuration or an error
// if parsing fails.
func NewServerConfig() (*ServerConfig, error) {
	envs, err := parseServerEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to create the server configuration: %w", err)
	}

	var privateKey *rsa.PrivateKey
	// Setup public key from the provided path.
	if privateKey, err = setPrivateKey(envs.PrivateKeyPath); err != nil {
		return nil, fmt.Errorf("failed to setup private key: %w", err)
	}

	cfg := &ServerConfig{
		Envs:            envs,
		ShutdownTimeout: defaultShutdownTimeout,
		PrivateKey:      privateKey,
	}

	return cfg, nil
}

// setPrivateKey sets up the private key for encryption.
//
//nolint:dupl // Will be refactored in the future
func setPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	if privateKeyPath == "" {
		return nil, ErrPrivateKeyPathNotSet
	}

	if !utils.VerifyFileExists(privateKeyPath) {
		if err := helpers.GenerateKeyPair(privateKeyPath); err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}
