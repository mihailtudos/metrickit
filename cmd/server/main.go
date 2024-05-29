package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/mihailtudos/metrickit/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
)

// TODO(SSH): I would move it closer to the spot where it is actually used
//
//go:embed templates
var templatesFs embed.FS

func main() {
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to provide server config: " + err.Error())
	}

	if err = run(appConfig); err != nil {
		log.Fatalf("failed to initialize the mem store: %s", err.Error())
	}
}

func run(cfg *config.ServerConfig) error {
	store, err := storage.NewMemStorage(cfg)
	if err != nil {
		if cfg.Log != nil {
			cfg.Log.ErrorContext(context.Background(), "failed to initialize the mem")
		}
		return fmt.Errorf("failed to setup the memstore: %w", err)
	}
	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(server.NewService(repos, cfg.Log), cfg.Log, templatesFs)

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
