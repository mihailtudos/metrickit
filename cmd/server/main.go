package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service"
)

func main() {
	appConfig := config.NewServerConfig()
	run(appConfig)
}

func run(cfg *config.ServerConfig) {
	store := storage.NewMemStorage()

	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(service.NewService(repos, cfg.Log), cfg.Log)

	cfg.Log.InfoContext(context.Background(), "running server 🔥", slog.String("port", cfg.Address))

	if err := http.ListenAndServe(cfg.Address, h.InitHandlers()); err != nil {
		cfg.Log.ErrorContext(context.Background(), "server failed to start: ", slog.String("error", err.Error()))
	}
}
