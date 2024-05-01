package main

import (
	"fmt"
	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service"
	"log"
	"net/http"
)

func main() {
	parseFlags()
	appConfig := config.NewAppConfig(addr.String())
	run(appConfig)
}

func run(cfg config.AppConfig) {
	store := storage.NewMemStorage()

	repos := repositories.NewRepository(store)
	h := handlers.NewHandler(service.NewService(repos, cfg.Log), cfg.Log)

	fmt.Printf("running server 🔥 on port: %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", cfg.Address), h.InitHandlers()))
}
