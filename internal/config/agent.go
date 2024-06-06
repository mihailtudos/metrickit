package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

const defaultReportInterval = 10
const defaultPoolInterval = 2

type AgentEnvs struct {
	Log            *slog.Logger
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type envAgentConfig struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func NewAgentConfig() (*AgentEnvs, error) {
	logger := NewLogger(os.Stdout, defaultLogLevel)

	envs, err := parseAgentEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to create agent config: %w", err)
	}

	return &AgentEnvs{
		Log:            logger,
		ServerAddr:     envs.ServerAddr,
		PollInterval:   time.Duration(envs.PollInterval) * time.Second,
		ReportInterval: time.Duration(envs.ReportInterval) * time.Second,
	}, nil
}

func parseAgentEnvs() (*envAgentConfig, error) {
	envConfig := &envAgentConfig{
		PollInterval:   defaultPoolInterval,
		ReportInterval: defaultReportInterval,
		ServerAddr:     fmt.Sprintf("%s:%d", defaultAddress, defaultPort),
	}

	flag.StringVar(&envConfig.ServerAddr, "a", envConfig.ServerAddr,
		"server address - usage: ADDRESS:PORT")
	flag.IntVar(&envConfig.PollInterval, "p", envConfig.PollInterval,
		"sets the frequency of polling the metrics in seconds")
	flag.IntVar(&envConfig.ReportInterval, "r", envConfig.ReportInterval,
		"sets the frequency of sending metrics to the server in seconds")

	flag.Parse()

	if err := env.Parse(envConfig); err != nil {
		return nil, fmt.Errorf("agent configs: %w", err)
	}

	return envConfig, nil
}
