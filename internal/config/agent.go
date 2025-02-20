// Package config provides configuration settings for the agent application.
//
// It supports configuration through environment variables, command-line flags,
// and default settings. The package allows setting up the agent's environment
// variables, including logging, server address, secret key, rate limits, and
// polling intervals.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	env11 "github.com/caarlos0/env/v11"
	"github.com/spf13/viper"

	"github.com/mihailtudos/metrickit/internal/logger"
	"github.com/mihailtudos/metrickit/internal/utils"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

const (
	// Default values for various configuration settings.
	defaultReportInterval = 10 // Default interval for reporting metrics, in seconds.
	defaultPoolInterval   = 2  // Default interval for polling metrics, in seconds.
	defaultRateLimit      = 10 // Default rate limit for concurrent operations.
)

// AgentEnvs represents the agent's runtime configuration settings.
type AgentEnvs struct {
	PublicKey      *rsa.PublicKey // Public key for encryption, configurable via environment variable "CRYPTO_KEY".
	Log            *slog.Logger   // Logger used by the agent.
	ServerAddr     string         // Address of the server to which metrics are sent.
	Key            string         // Secret key used for signing data.
	GRPCAddress    string         // gRPC server address, configurable via environment variable "GRPC_ADDRESS".
	RateLimit      int            // Maximum number of concurrent goroutines.
	PollInterval   time.Duration  // Interval between metric polling operations.
	ReportInterval time.Duration  // Interval between sending metrics to the server.
}

// envAgentConfig is a struct for parsing environment variables into agent configuration settings.
type envAgentConfig struct {
	ServerAddr  string `env:"ADDRESS" json:"address"`
	GRPCAddress string `env:"GRPC_ADDRESS" json:"grpc_address"`
	// Server address, configurable via environment variable "ADDRESS".
	LogLevel string `env:"LOG_LEVEL"`
	// Logging level, configurable via environment variable "LOG_LEVEL".
	Key           string `env:"KEY"`
	PublicKeyPath string `env:"CRYPTO_KEY" json:"crypto_key"` // Public key file path, configurable via env "CRYPTO_KEY".
	// Rate limit, configurable via environment variable "RATE_LIMIT".
	ConfigFilePath string `env:"CONFIG"`
	// Secret key, configurable via environment variable "KEY".
	PollInterval int `env:"POLL_INTERVAL" json:"poll_interval"`
	// Polling interval in seconds, configurable via environment variable "POLL_INTERVAL".
	ReportInterval int `env:"REPORT_INTERVAL" json:"report_interval"`
	// Reporting interval in seconds, configurable via environment variable "REPORT_INTERVAL".
	RateLimit int `env:"RATE_LIMIT"`
}

// NewAgentConfig creates a new AgentEnvs instance by parsing environment variables
// and command-line flags. It sets up default values, overrides them with environment
// variables if provided, and applies command-line flag values.
//
// Returns:
//   - *AgentEnvs: A pointer to the populated AgentEnvs configuration.
//   - error: An error if configuration parsing fails.
func NewAgentConfig() (*AgentEnvs, error) {
	envs, err := parseAgentEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to create agent config: %w", err)
	}

	// Initialize the logger with the specified log level.
	l, err := logger.NewLogger(os.Stdout, envs.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("agent logger: %w", err)
	}

	var publicKey *rsa.PublicKey
	// Setup public key from the provided path.
	if publicKey, err = setupPublicKey(envs.PublicKeyPath); err != nil {
		return nil, fmt.Errorf("failed to setup public key: %w", err)
	}

	return &AgentEnvs{
		Log:            l,
		ServerAddr:     envs.ServerAddr,
		PollInterval:   time.Duration(envs.PollInterval) * time.Second,
		ReportInterval: time.Duration(envs.ReportInterval) * time.Second,
		Key:            envs.Key,
		RateLimit:      envs.RateLimit,
		PublicKey:      publicKey,
		GRPCAddress:    envs.GRPCAddress,
	}, nil
}

// parseAgentEnvs reads environment variables and command-line flags to populate
// an envAgentConfig instance. It applies default values first, then overrides
// them with environment variables, and finally with command-line flags.
//
// Returns:
//   - *envAgentConfig: A pointer to the populated envAgentConfig struct.
//   - error: An error if environment parsing fails.
func parseAgentEnvs() (*envAgentConfig, error) {
	envConfig := &envAgentConfig{
		LogLevel:       DefaultLogLevel,                                   // Default log level.
		PollInterval:   defaultPoolInterval,                               // Default polling interval.
		ReportInterval: defaultReportInterval,                             // Default reporting interval.
		RateLimit:      defaultRateLimit,                                  // Default rate limit.
		ServerAddr:     fmt.Sprintf("%s:%d", defaultAddress, defaultPort), // Default server address.
	}

	// Command-line flags override default values and environment variables.
	flag.StringVar(&envConfig.LogLevel, "ll",
		envConfig.LogLevel, "log level")
	flag.StringVar(&envConfig.ConfigFilePath, "config",
		"", "agent json configuration file path")
	flag.StringVar(&envConfig.ServerAddr, "a",
		envConfig.ServerAddr, "server address - usage: ADDRESS:PORT")
	flag.StringVar(&envConfig.Key, "k",
		envConfig.Key,
		"sets the secret key used for signing data")
	flag.IntVar(&envConfig.PollInterval, "p",
		envConfig.PollInterval,
		"sets the frequency of polling the metrics in seconds")
	flag.IntVar(
		&envConfig.ReportInterval, "r",
		envConfig.ReportInterval,
		"sets the frequency of sending metrics to the server in seconds")
	flag.IntVar(&envConfig.RateLimit, "l",
		envConfig.RateLimit,
		"rate limit, max goroutines to run at a time")
	flag.StringVar(&envConfig.PublicKeyPath, "crypto-key",
		envConfig.PublicKeyPath,
		"path to the public key file")
	flag.StringVar(&envConfig.GRPCAddress, "grpc-addr",
		"",
		"sets the address for gRPC communication")

	flag.Parse()

	// Parse environment variables into the envConfig struct.
	if err := env11.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("agent configs: %w", err)
	}

	// Override default values with config variables.
	if envConfig.ConfigFilePath != "" {
		viper.SetConfigName("agent") // name of the file without extension
		viper.SetConfigType("json")  // specify the file type
		viper.AddConfigPath(envConfig.ConfigFilePath)

		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read json config file: %w", err)
		}

		if envConfig.ServerAddr == "" {
			envConfig.ServerAddr = viper.GetString("address")
		}

		utils.Replace(&envConfig.ServerAddr, viper.GetString("address"))
		utils.Replace(&envConfig.PublicKeyPath, viper.GetString("crypto_key"))
		utils.Replace(&envConfig.PollInterval, int(viper.GetDuration("poll_interval").Seconds()))
		utils.Replace(&envConfig.ReportInterval, int(viper.GetDuration("report_interval").Seconds()))
		utils.Replace(&envConfig.GRPCAddress, viper.GetString("grpc_address"))
	}

	fmt.Printf("%+v", envConfig)
	return envConfig, nil
}

// setupPublicKey sets up the public key for encryption and decryption.
//
//nolint:dupl // This is a duplicate but servers a different purpose.
func setupPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	if publicKeyPath == "" {
		return nil, ErrPublicKeyPathNotProvided
	}

	if !utils.VerifyFileExists(publicKeyPath) {
		if err := helpers.GenerateKeyPair(publicKeyPath); err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
	}

	// Read and parse public key
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	p, _ := pem.Decode(publicKeyBytes)
	publicKey, err := x509.ParsePKCS1PublicKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return publicKey, nil
}
