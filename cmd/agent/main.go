package main

import (
	"time"

	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service"
)

func main() {
	agentCfg := config.NewAgentConfig()
	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore)
	metricsService := service.NewAgentService(metricsRepo, agentCfg.Log)

	poolTicker := time.NewTicker(agentCfg.PollInterval)
	defer poolTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-poolTicker.C:
			metricsService.Collect()
		case <-reportTicker.C:
			metricsService.Send(agentCfg.ServerAddr)
			metricsService.Clear()
		}
	}
}
