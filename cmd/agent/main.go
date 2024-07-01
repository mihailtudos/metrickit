package main

import (
	"context"
	"github.com/mihailtudos/metrickit/internal/worker"
	"os"
	"os/signal"
	"syscall"
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
	// Create worker pool with the specified rate limit
	workerPool := worker.NewWorkerPool(agentCfg.RateLimit)
	workerPool.Run()

	pollTicker := time.NewTicker(agentCfg.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	//originalReportInterval := agentCfg.ReportInterval

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	go func() {
		for {
			select {
			case <-pollTicker.C:
				if err := metricsService.MetricsService.Collect(); err != nil {
					agentCfg.Log.ErrorContext(context.Background(),
						"failed to collect the metrics: ",
						helpers.ErrAttr(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-reportTicker.C:
			task := &agent.SendMetricsTask{
				Service:    metricsService,
				ServerAddr: agentCfg.ServerAddr,
			}
			workerPool.AddTask(task)
		case <-ctx.Done():
			workerPool.Stop()
			workerPool.Wait()
			return
		}
	}
}
