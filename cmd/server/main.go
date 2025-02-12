package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/database"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/handlers"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/logger"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/mihailtudos/metrickit/internal/utils"
	"github.com/mihailtudos/metrickit/pkg/helpers"

	_ "net/http/pprof"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const timeToShutDown = 5

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
	// Output the build information
	fmt.Println(utils.BuildTagsFormatedString(buildVersion, buildDate, buildCommit))

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

// run function starts the server and handles the server's lifecycle.
func (app *ServerApp) run(ctx context.Context) error {
	app.logger.DebugContext(ctx, "provided_config",
		slog.String("ServerAddress", app.cfg.Envs.Address),
		slog.String("StorePath", app.cfg.Envs.StorePath),
		slog.String("LogLevel", app.cfg.Envs.LogLevel),
		slog.Int("StoreInterval", app.cfg.Envs.StoreInterval),
		slog.Bool("ReStore", app.cfg.Envs.ReStore),
		slog.Bool("Secret", app.cfg.Envs.Key != ""))

	// Initialize storage
	store, err := storage.NewStorage(app.db, app.logger, app.cfg.Envs.StoreInterval, app.cfg.Envs.StorePath)
	if err != nil {
		app.logger.ErrorContext(ctx, "failed to initialize storage")
		return fmt.Errorf("failed to setup storage: %w", err)
	}
	defer func() {
		app.logger.DebugContext(ctx, "shutting down storage")
		if err := store.Close(ctx); err != nil {
			app.logger.ErrorContext(ctx, "failed to close storage", helpers.ErrAttr(err))
		}
	}()

	// Initialize repositories and services
	repos := repositories.NewRepository(store)
	service := server.NewMetricsService(repos, app.logger)
	serverHandlers := handlers.NewHandler(service, app.logger, app.db, app.cfg.Envs.Key,
		app.cfg.PrivateKey, app.cfg.TrustedSubnet)

	// Start HTTP server
	srv := &http.Server{
		Addr:    app.cfg.Envs.Address,
		Handler: handlers.Router(app.logger, serverHandlers),
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		app.logger.DebugContext(ctx, "starting server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.ErrorContext(ctx, "server error", helpers.ErrAttr(err))
		}
	}()

	// Wait for termination signal
	sig := <-signalCh
	app.logger.InfoContext(ctx, "received signal, shutting down server",
		slog.String("signal", sig.String()))

	// Shutdown the server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, timeToShutDown*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		app.logger.ErrorContext(ctx, "failed to shutdown server gracefully", helpers.ErrAttr(err))
	}

	// Additional cleanup for the database connection pool
	if app.db != nil {
		app.logger.DebugContext(ctx, "shutting down the db connection pool")
		app.db.Close()
	}

	app.logger.InfoContext(ctx, "server stopped successfully")
	return nil
}
