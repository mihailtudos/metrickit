package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/mihailtudos/metrickit/pkg/flags"
	"log"
	"time"
)

const DefaultReportInterval = 10
const DefaultPoolInterval = 2

var serverAddr *flags.ServerAddr
var poolIntervalInSeconds *flags.DurationFlag
var reportIntervalInSeconds *flags.DurationFlag

type AgentConfig struct {
	PollInterval   *time.Duration `env:"POLL_INTERVAL"`
	ReportInterval *time.Duration `env:"REPORT_INTERVAL"`
	ServerAddr     string         `env:"ADDRESS"`
}

func NewAgentConfig() *AgentConfig {
	parseFlags()

	return &AgentConfig{
		PollInterval:   poolIntervalInSeconds.GetDuration(),
		ReportInterval: reportIntervalInSeconds.GetDuration(),
		ServerAddr:     serverAddr.String(),
	}
}

func parseFlags() {
	var cfg AgentConfig
	if err := env.Parse(&cfg); err != nil {
		log.Panic("failed to parse env vars")
	}

	serverAddr = flags.NewServerAddressFlag(DefaultAddress, DefaultPort)
	poolIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultPoolInterval)
	reportIntervalInSeconds = flags.NewDurationFlag(time.Second, DefaultReportInterval)

	_ = flag.Value(serverAddr)

	flag.Var(serverAddr, "a", "server address - usage: ADDRESS:PORT")
	flag.Var(poolIntervalInSeconds, "p", "frequency of polling the metrics in seconds e.g. -p=2s")
	flag.Var(reportIntervalInSeconds, "r", "frequency of sending metrics to the server in seconds e.g. -r=4s")

	flag.Parse()

	host, port, err := splitAddressParts(cfg.ServerAddr)
	if err == nil {
		serverAddr = flags.NewServerAddressFlag(host, port)
	}

	setConfig(cfg.ReportInterval, reportIntervalInSeconds)
	setConfig(cfg.PollInterval, poolIntervalInSeconds)
}

func setConfig(interval *time.Duration, config *flags.DurationFlag) {
	if interval != nil {
		t, err := validateAndConvertToSeconds(*interval)
		if err == nil {
			config.Length = t
		}
	}
}

func validateAndConvertToSeconds(duration time.Duration) (int, error) {
	durationInSeconds := int(duration.Seconds())
	if durationInSeconds < 0 {
		return 0, fmt.Errorf("ReportInterval must be a positive duration")
	}

	return durationInSeconds, nil
}
