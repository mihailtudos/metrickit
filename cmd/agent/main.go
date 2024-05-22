package main

import (
	"context"
	"log"
	"time"

	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/agent"
)

func main() {
	agentCfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatal("failed to get agent configurations: " + err.Error())
	}
	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore)
	metricsService := agent.NewAgentService(metricsRepo, agentCfg.Log)

	poolTicker := time.NewTicker(agentCfg.PollInterval)
	defer poolTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-poolTicker.C:
			if err := metricsService.Collect(); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to collect the metrics: "+err.Error())
			}
		case <-reportTicker.C:
			if err := metricsService.Send(agentCfg.ServerAddr); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to publish the metrics: "+err.Error())
			}
			metricsStore.Clear()
		}
	}
}
