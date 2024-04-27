package main

import (
	"fmt"
	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/services"
	"log"
	"net/http"
)

const port = "8080"

func main() {
	appConfig := config.NewAppConfig(port)
	run(appConfig)
}

func run(cfg config.AppConfig) {
	store := storage.NewMemStorage()

	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(services.NewService(repos, cfg.Log), cfg.Log)

	fmt.Printf("running server 🔥 on port: %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Address), h.InitHandlers()))
}
