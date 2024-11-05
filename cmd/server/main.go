package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/database"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/logger"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/mihailtudos/metrickit/pkg/helpers"

	_ "net/http/pprof"

	"github.com/jackc/pgx/v5/pgxpool"
)

//nolint:godot // this comment is part of the Swagger documentation
// @BasePath  /
// @Title Metrics API
// @Description Metrics service for monitoring, retrieving, and managing metric data.
// This API allows for querying the values of various metrics and supports updating metric values.
// @Version 1.0
// @Contact.email support@example.com.
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /
// @Tag.name Info
// @Tag.description "Endpoints for retrieving the status and information of the service."
// @Tag.name Metric Storage
// @Tag.description "Endpoints for managing and accessing metric data stored in the service."

// ServerApp holds the required dependecies for the metrics service.
type ServerApp struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	cfg    *config.ServerConfig
}

func main() {
	// appConfig - holds a pointer to the server configurations
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to initiate server config: " + err.Error())
	}

	// initializing a new logger
	newLogger, err := logger.NewLogger(os.Stdout, appConfig.Envs.LogLevel)
	if err != nil {
		log.Fatal("failed to initiate server logger: " + err.Error())
	}

	// building the server
	app := ServerApp{logger: newLogger, cfg: appConfig}

	// initializing a DB conn pool if the DSN was provided
	ctx := context.Background()
	if app.cfg.Envs.D3SN != "" {
		db, e := database.InitPostgresDB(ctx, app.cfg.Envs.D3SN, app.logger)
		if e != nil {
			app.logger.ErrorContext(ctx, "failed to init db", helpers.ErrAttr(e))
			log.Fatal("failed to init db: " + e.Error())
		}

		app.db = db
	}

	// running the server in the main thread
	if err = app.run(ctx); err != nil {
		app.logger.ErrorContext(context.Background(),
			"failed to run the server",
			helpers.ErrAttr(err))
		log.Fatal(err)
	}
}

func (app *ServerApp) run(ctx context.Context) error {
	app.logger.DebugContext(ctx, "provided_config",
		slog.String("ServerAddress", app.cfg.Envs.Address),
		slog.String("StorePath", app.cfg.Envs.StorePath),
		slog.String("LogLevel", app.cfg.Envs.LogLevel),
		slog.Int("StoreInterval", app.cfg.Envs.StoreInterval),
		slog.Bool("ReStore", app.cfg.Envs.ReStore),
		slog.Bool("Secret", app.cfg.Envs.Key != ""))

	store, err := storage.NewStorage(app.db,
		app.logger,
		app.cfg.Envs.StoreInterval,
		app.cfg.Envs.StorePath)
	if err != nil {
		app.logger.ErrorContext(ctx, "failed to initialize the mem")
		return fmt.Errorf("failed to setup the memstore: %w", err)
	}

	defer func() {
		if err = store.Close(ctx); err != nil {
			app.logger.ErrorContext(ctx,
				"failed to close the DB connection",
				helpers.ErrAttr(err))
		}
	}()

	repos := repositories.NewRepository(store)
	service := server.NewMetricsService(repos, app.logger)
	router := handlers.NewHandler(service, app.logger, app.db, app.cfg.Envs.Key)

	app.logger.DebugContext(context.Background(), "running server 🔥",
		slog.String("address", app.cfg.Envs.Address))
	srv := &http.Server{
		Addr:    app.cfg.Envs.Address,
		Handler: router,
	}

	if err = srv.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	return nil
}
