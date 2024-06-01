package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		slog.String("ServerAddress", cfg.Address),
		slog.String("StorePath", cfg.StorePath),
		slog.String("LogLevel", cfg.LogLevel),
		slog.Int("StoreInterval", cfg.StoreInterval),
		slog.Bool("ReStore", cfg.ReStore))

	store, err := storage.NewStorage(cfg)
	if err != nil {
		cfg.Log.ErrorContext(context.Background(), "failed to initialize the mem")
		return fmt.Errorf("failed to setup the memstore: %w", err)
	}
	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(server.NewMetricsService(repos, cfg.Log), cfg.Log)

	cfg.Log.DebugContext(context.Background(), "running server 🔥 on port: "+cfg.Address)
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: h.InitHandlers(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on %s: %v\n", cfg.Address, err)
		}
	}()

	<-quit
	cfg.Log.DebugContext(context.Background(), "shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()
	err = store.Close()
	if err != nil {
		cfg.Log.ErrorContext(ctx, "failed to close the file", err)
	}

	if err = srv.Shutdown(ctx); err != nil {
		cfg.Log.ErrorContext(ctx, "server forced to shutdown", err)
	}

	cfg.Log.DebugContext(context.Background(), "server exiting")
	return nil
}
