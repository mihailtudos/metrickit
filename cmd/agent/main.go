package main

import (
	"context"
	"os"

	"github.com/mihailtudos/metrickit/internal/agent"
	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

func main() {
	agentCfg, err := config.NewAgentConfig()
	if err != nil {
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to get agent configurations: ",
			helpers.ErrAttr(err),
		)
		os.Exit(1)
	}

	if err := agent.RunAgent(agentCfg); err != nil {
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to start the agent: ",
			helpers.ErrAttr(err),
		)
		os.Exit(1)
	}
}
