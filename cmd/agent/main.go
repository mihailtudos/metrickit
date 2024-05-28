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
		// TODO(SSH):
		// 1. as you use slog as the logger in the middleware, it makes sense to use it exclusively in the project
		// so no other packages (including `log` from stdlib) are used anywhere
		// 2. log.Fatalf("failed to get agent configurations: %v", err)
		log.Fatal("failed to get agent configurations: " + err.Error())
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
			if err := metricsService.Collect(); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to collect the metrics: "+err.Error())
			}
		case <-reportTicker.C:
			if err := metricsService.SendJSONMetric(agentCfg.ServerAddr); err != nil {
				agentCfg.Log.ErrorContext(context.Background(), "failed to publish the metrics: "+err.Error())
			}
			metricsStore.Clear()
		}
	}
}
