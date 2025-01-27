package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mihailtudos/metrickit/internal/agent"
	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/utils"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Output the build information
	fmt.Println(utils.BuildTagsFormatedString(buildVersion, buildDate, buildCommit))

	agentCfg, err := config.NewAgentConfig()
	if err != nil {
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to get agent configurations: ",
			helpers.ErrAttr(err),
		)
		log.Fatal(err.Error())
	}

	if err = agent.RunAgent(agentCfg); err != nil {
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to start the agent: ",
			helpers.ErrAttr(err),
		)
		log.Fatal(err.Error())
	}
}
