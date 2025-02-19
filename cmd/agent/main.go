package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/agent"
	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	agentCfg, err := config.NewAgentConfig()
	if err != nil {
		log.Println(
			"failed to get agent configurations: ",
			helpers.ErrAttr(err),
		)
		log.Fatal(err.Error())
	}

	// Output the build information
	agentCfg.Log.InfoContext(context.Background(), "agent built info",
		slog.String("version", buildVersion),
		slog.String("date", buildDate),
		slog.String("commit", buildCommit))

	if err = agent.RunAgent(agentCfg); err != nil {
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to start the agent: ",
			helpers.ErrAttr(err),
		)
		log.Fatal(err.Error())
	}
}
