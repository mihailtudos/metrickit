package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
)

func main() {
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to provide server config: " + err.Error())
	}

	if err = run(appConfig); err != nil {
		appConfig.Log.ErrorContext(context.Background(), "failed to run the server: "+err.Error())
		os.Exit(1)
	}
}

func run(cfg *config.ServerConfig) error {
	cfg.Log.DebugContext(context.Background(), "provided config",
		slog.String("ServerAddress", cfg.Envs.Address),
		slog.String("StorePath", cfg.Envs.StorePath),
		slog.String("LogLevel", cfg.Envs.LogLevel),
		slog.Int("StoreInterval", cfg.Envs.StoreInterval),
		slog.Bool("ReStore", cfg.Envs.ReStore))

	store, err := storage.NewStorage(cfg)
	if err != nil {
		cfg.Log.ErrorContext(context.Background(), "failed to initialize the mem")
		return fmt.Errorf("failed to setup the memstore: %w", err)
	}
	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(server.NewMetricsService(repos, cfg.Log), cfg.Log)

	cfg.Log.DebugContext(context.Background(), "running server 🔥 on port: "+cfg.Envs.Address)
	srv := &http.Server{
		Addr:    cfg.Envs.Address,
		Handler: h.InitHandlers(),
	}

	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("could not listen on %s: %v\n", cfg.Envs.Address, err)
	}

	return nil
}
