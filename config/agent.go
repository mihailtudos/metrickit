package config

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/pkg/flags"
)

const DefaultReportInterval = 10
const DefaultPoolInterval = 2

var serverAddr *flags.ServerAddr
var poolIntervalInSeconds *flags.DurationFlag
var reportIntervalInSeconds *flags.DurationFlag

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

func NewAgentConfig() *AgentConfig {
	parseFlags()

	return &AgentConfig{
		PollInterval:   poolIntervalInSeconds.GetDuration(),
		ReportInterval: reportIntervalInSeconds.GetDuration(),
		ServerAddr:     serverAddr.String(),
		Log:            slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func parseFlags() {
	var cfg EnvAgentConfig
	if err := env.Parse(&cfg); err != nil {
		log.Panic("failed to parse env vars: ", err.Error())
	}

	serverAddr = flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
	poolIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultPoolInterval)
	reportIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultReportInterval)

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
}

func setConfig(interval *int, config *flags.DurationFlag) {
	if interval != nil {
		config.Length = *interval
	}
}
