package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/agent"
	"github.com/mihailtudos/metrickit/internal/worker"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore, agentCfg.Log)
	metricsService := agent.NewAgentService(metricsRepo, agentCfg.Log, &agentCfg.Key)

	workerPool := worker.NewWorkerPool(agentCfg.RateLimit)

	pollTicker := time.NewTicker(agentCfg.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-pollTicker.C:
				if err = metricsService.MetricsService.Collect(); err != nil {
					agentCfg.Log.ErrorContext(ctx,
						"failed to collect the metrics: ",
						helpers.ErrAttr(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	workerPool.Run(ctx)
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-reportTicker.C:
				task := &agent.SendMetricsTask{
					Service:    metricsService,
					ServerAddr: agentCfg.ServerAddr,
					Log:        agentCfg.Log,
				}
				workerPool.AddTask(task)
			case <-ctx.Done():
				workerPool.Wait()
				return
			}
		}
	}()

	wg.Wait()
}
