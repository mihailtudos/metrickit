package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/agent"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

func main() {
	agentCfg, err := config.NewAgentConfig()
	if err != nil {
		fmt.Println(err)
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to get agent configurations: ",
			helpers.ErrAttr(err),
		)
		os.Exit(-1)
	}

	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore, agentCfg.Log)
	metricsService := agent.NewAgentService(metricsRepo, agentCfg.Log)

	poolTicker := time.NewTicker(agentCfg.PollInterval)
	defer poolTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-poolTicker.C:
			if err := metricsService.MetricsService.Collect(); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to collect the metrics: ", helpers.ErrAttr(err))
			}
		case <-reportTicker.C:
			if err := metricsService.MetricsService.Send(agentCfg.ServerAddr); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to publish the metrics: "+err.Error())
			}
			metricsStore.Clear()
		}
	}
}
