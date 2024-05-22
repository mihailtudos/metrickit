package config

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/pkg/flags"
)

const DefaultReportInterval = 10
const DefaultPoolInterval = 2

type AgentConfig struct {
	Log            *slog.Logger
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type EnvAgentConfig struct {
	PollInterval   *int    `env:"POLL_INTERVAL"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	ServerAddr     *string `env:"ADDRESS"`
}

func NewAgentConfig() (*AgentConfig, error) {
	cfg := AgentConfig{}

	err := parseFlags(&cfg)
	if err != nil {
		return nil, errors.New("failed to create agent config: " + err.Error())
	}

	cfg.Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &cfg, nil
}

func parseFlags(agentCfg *AgentConfig) error {
	var cfg EnvAgentConfig
	if err := env.Parse(&cfg); err != nil {
		return errors.New("failed to parse env vars: " + err.Error())
	}

	serverAddr := flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
	poolIntervalInSeconds := flags.NewDurationFlag(time.Second, DefaultPoolInterval)
	reportIntervalInSeconds := flags.NewDurationFlag(time.Second, DefaultReportInterval)

	_ = flag.Value(serverAddr)

	flag.Var(serverAddr, "a", "server address - usage: ADDRESS:PORT")
	flag.Var(poolIntervalInSeconds, "p", "sets the frequency of polling the metrics in seconds e.g. -p=2")
	flag.Var(reportIntervalInSeconds, "r", "sets the frequency of sending metrics to the server in seconds e.g. -r=4")

	flag.Parse()

	host, port, err := splitAddressParts(cfg.ServerAddr)
	if err == nil {
		serverAddr = flags.NewServerAddressFlag(host, port)
	}

	setConfig(cfg.ReportInterval, reportIntervalInSeconds)
	setConfig(cfg.PollInterval, poolIntervalInSeconds)

	agentCfg.PollInterval = poolIntervalInSeconds.GetDuration()
	agentCfg.ReportInterval = reportIntervalInSeconds.GetDuration()
	agentCfg.ServerAddr = serverAddr.String()

	return nil
}

func setConfig(interval *int, config *flags.DurationFlag) {
	if interval != nil {
		config.Length = *interval
	}
}
