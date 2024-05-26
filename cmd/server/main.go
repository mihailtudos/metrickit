package main

import (
	"context"
	"embed"
	"log"
	"net/http"

	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
)

//go:embed templates
var templatesFs embed.FS

func main() {
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to provide server config: " + err.Error())
	}

	run(appConfig)
}

func run(cfg *config.ServerConfig) {
	store := storage.NewMemStorage()

	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(server.NewService(repos, cfg.Log), cfg.Log, templatesFs)

	cfg.Log.DebugContext(context.Background(), "running server 🔥 on port: "+cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, h.InitHandlers()))
}
