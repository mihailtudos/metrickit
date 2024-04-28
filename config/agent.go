package config

import "time"

type AgentCfg struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddr     string
}

func NewAgentCfg(pollItv time.Duration, reportItv time.Duration, serverAddr string) *AgentCfg {
	return &AgentCfg{
		PollInterval:   pollItv,
		ReportInterval: reportItv,
		ServerAddr:     serverAddr,
	}
}
