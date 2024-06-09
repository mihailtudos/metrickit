package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mihailtudos/metrickit/pkg/helpers"
	"log"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
)

func main() {
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to initiate server config: " + err.Error())
	}

	if err = run(appConfig); err != nil {
		appConfig.Log.ErrorContext(context.Background(), "failed to run the server: "+err.Error())
		log.Fatal(err)
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
	defer func() {
		if err := store.Close(); err != nil {
			cfg.Log.ErrorContext(context.Background(),
				"failed to close the DB connection",
				helpers.ErrAttr(err))
		}
	}()

	if err != nil {
		cfg.Log.ErrorContext(context.Background(), "failed to initialize the mem")
		return fmt.Errorf("failed to setup the memstore: %w", err)
	}
	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(server.NewMetricsService(repos, cfg.Log), cfg.Log, cfg.DB)

	cfg.Log.DebugContext(context.Background(), "running server 🔥", slog.String("address", cfg.Envs.Address))
	srv := &http.Server{
		Addr:    cfg.Envs.Address,
		Handler: h.InitHandlers(),
	}

	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	return nil
}
