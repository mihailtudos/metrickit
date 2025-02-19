// Package agent is responsible for running the agent that manages the core
// logic and configuration for metric collection and reporting.
//
// The agent periodically collects metrics and reports them to a specified
// server. It supports graceful shutdown and leverages worker pools for
// handling background tasks. The configuration settings dictate the polling
// and reporting intervals, as well as the rate limits for task execution.
package agent

import (
	"context"
	"log/slog"
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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RunAgent initializes and starts the agent with the provided configuration.
//
// The function sets up background tasks for collecting and reporting metrics
// at specified intervals. It also initializes the necessary services and
// repositories. The agent handles termination signals to gracefully shut down
// and ensures all background tasks are completed before exiting.
//
// Parameters:
//   - agentCfg: The configuration settings for the agent, including intervals,
//     logging, rate limits, and server address.
//
// Returns:
//   - error: If an error occurs during the agent's operation, it is returned.
//     If the agent runs successfully without issues, nil is returned.
func RunAgent(agentCfg *config.AgentEnvs) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Initialize storage and services for metrics collection.
	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore, agentCfg.Log)

	var conn *grpc.ClientConn
	if agentCfg.GRPCAddress != "" {
		var err error
		conn, err = grpc.NewClient(agentCfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			agentCfg.Log.ErrorContext(ctx,
				"Failed to create grpc connection",
				helpers.ErrAttr(err),
			)
			return err
		}

		defer func() {
			if errGRPCConn := conn.Close(); errGRPCConn != nil {
				agentCfg.Log.Error("Failed to close grpc connection", helpers.ErrAttr(errGRPCConn))
			}
		}()
		agentCfg.Log.DebugContext(ctx, "GRPC address provided", slog.String("address", agentCfg.GRPCAddress))
	} else {
		agentCfg.Log.DebugContext(ctx, "GRPC address is not provided")
	}

	metricsService := agent.NewAgentService(
		metricsRepo,
		agentCfg.Log,
		&agentCfg.Key,
		agentCfg.PublicKey,
		conn,
	)

	// Set up a worker pool with rate limiting.
	workerPool := worker.NewWorkerPool(agentCfg.RateLimit)

	// Create timers for polling and reporting intervals.
	pollTicker := time.NewTicker(agentCfg.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(agentCfg.ReportInterval)
	defer reportTicker.Stop()

	// Set up a channel to listen for termination signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigChan
		cancel()
	}()

	// Start a background task for collecting metrics at regular intervals.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-pollTicker.C:
				if err := metricsService.MetricsService.Collect(); err != nil {
					agentCfg.Log.ErrorContext(ctx,
						"failed to collect the metrics: ",
						helpers.ErrAttr(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start the worker pool to execute background tasks.
	workerPool.Run(ctx)
	wg.Add(1)

	// Start a background task for reporting metrics at regular intervals.
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

	// Wait for all background tasks to complete before exiting.
	wg.Wait()

	return nil
}
