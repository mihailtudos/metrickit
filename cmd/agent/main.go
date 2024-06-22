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
		agentCfg.Log.ErrorContext(context.Background(),
			"failed to get agent configurations: ",
			helpers.ErrAttr(err),
		)
		os.Exit(1)
	}

	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore, agentCfg.Log)
	metricsService := agent.NewAgentService(metricsRepo, agentCfg.Log, &agentCfg.Key)

	pollTicker := time.NewTicker(agentCfg.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	originalReportInterval := agentCfg.ReportInterval

	for {
		select {
		case <-pollTicker.C:
			if err = metricsService.MetricsService.Collect(); err != nil {
				agentCfg.Log.ErrorContext(context.Background(),
					"failed to collect the metrics: ",
					helpers.ErrAttr(err))
			}
		case <-reportTicker.C:
			retries := 3
			backoffIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
			success := false

			for attempt := range backoffIntervals {
				err = metricsService.MetricsService.Send(agentCfg.ServerAddr)
				if err == nil {
					metricsStore.Clear()
					reportTicker.Reset(originalReportInterval)
					success = true
					break
				}

				agentCfg.Log.ErrorContext(context.Background(),
					fmt.Sprintf("Attempt %d: failed to publish the metrics: %v", attempt+1, err),
					helpers.ErrAttr(err))

				if attempt < retries {
					time.Sleep(backoffIntervals[attempt])
				}
			}

			if !success {
				agentCfg.Log.ErrorContext(context.Background(),
					"All retry attempts failed. Exiting the program.",
					helpers.ErrAttr(err))
				return
			}
		}
	}
}
