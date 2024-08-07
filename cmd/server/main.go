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

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerApp struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	cfg    *config.ServerConfig
}

func main() {
	appConfig, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to initiate server config: " + err.Error())
	}

	newLogger, err := logger.NewLogger(os.Stdout, appConfig.Envs.LogLevel)
	if err != nil {
		log.Fatal("failed to initiate server logger: " + err.Error())
	}

	app := ServerApp{logger: newLogger, cfg: appConfig}

	ctx := context.Background()
	if app.cfg.Envs.D3SN != "" {
		db, err := database.InitPostgresDB(ctx, app.cfg.Envs.D3SN, app.logger)
		if err != nil {
			app.logger.ErrorContext(ctx, "failed to init db", helpers.ErrAttr(err))
			os.Exit(1)
		}

		app.db = db
	}

	if err = app.run(ctx); err != nil {
		app.logger.ErrorContext(context.Background(),
			"failed to run the server",
			helpers.ErrAttr(err))
		log.Fatal(err)
	}
}

func (app *ServerApp) run(ctx context.Context) error {
	app.logger.DebugContext(ctx, "provided config",
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
	h := handlers.NewHandler(service, app.logger, app.db, app.cfg.Envs.Key)

	app.logger.DebugContext(context.Background(), "running server 🔥",
		slog.String("address", app.cfg.Envs.Address))
	srv := &http.Server{
		Addr:    app.cfg.Envs.Address,
		Handler: h.InitHandlers(),
	}

	if err = srv.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	return nil
}
