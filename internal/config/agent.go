// Package config provides configuration settings for the agent application.
//
// It supports configuration through environment variables, command-line flags,
// and default settings. The package allows setting up the agent's environment
// variables, including logging, server address, secret key, rate limits, and
// polling intervals.
package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/internal/logger"
)

// Default values for various configuration settings.
const defaultReportInterval = 10 // Default interval for reporting metrics, in seconds.
const defaultPoolInterval = 2    // Default interval for polling metrics, in seconds.
const defaultRateLimit = 10      // Default rate limit for concurrent operations.

// AgentEnvs represents the agent's runtime configuration settings.
type AgentEnvs struct {
	Log            *slog.Logger  // Logger used by the agent.
	ServerAddr     string        // Address of the server to which metrics are sent.
	Key            string        // Secret key used for signing data.
	RateLimit      int           // Maximum number of concurrent goroutines.
	PollInterval   time.Duration // Interval between metric polling operations.
	ReportInterval time.Duration // Interval between sending metrics to the server.
}

// envAgentConfig is a struct for parsing environment variables into agent configuration settings.
type envAgentConfig struct {
	ServerAddr string `env:"ADDRESS"`
	// Server address, configurable via environment variable "ADDRESS".
	LogLevel string `env:"LOG_LEVEL"`
	// Logging level, configurable via environment variable "LOG_LEVEL".
	Key string `env:"KEY"`
	// Secret key, configurable via environment variable "KEY".
	PollInterval int `env:"POLL_INTERVAL"`
	// Polling interval in seconds, configurable via environment variable "POLL_INTERVAL".
	ReportInterval int `env:"REPORT_INTERVAL"`
	// Reporting interval in seconds, configurable via environment variable "REPORT_INTERVAL".
	RateLimit int `env:"RATE_LIMIT"`
	// Rate limit, configurable via environment variable "RATE_LIMIT".
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

	return &AgentEnvs{
		Log:            l,
		ServerAddr:     envs.ServerAddr,
		PollInterval:   time.Duration(envs.PollInterval) * time.Second,
		ReportInterval: time.Duration(envs.ReportInterval) * time.Second,
		Key:            envs.Key,
		RateLimit:      envs.RateLimit,
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
		LogLevel:       defaultLogLevel,                                   // Default log level.
		PollInterval:   defaultPoolInterval,                               // Default polling interval.
		ReportInterval: defaultReportInterval,                             // Default reporting interval.
		RateLimit:      defaultRateLimit,                                  // Default rate limit.
		ServerAddr:     fmt.Sprintf("%s:%d", defaultAddress, defaultPort), // Default server address.
	}

	// Command-line flags override default values and environment variables.
	flag.StringVar(&envConfig.LogLevel, "ll",
		envConfig.LogLevel, "log level")
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

	flag.Parse()

	// Parse environment variables into the envConfig struct.
	if err := env.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("agent configs: %w", err)
	}

	return envConfig, nil
}
