package handlers

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/internal/service/server"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed templates
var templatesFs embed.FS

type ServerHandler struct {
	services    *server.Service
	logger      *slog.Logger
	TemplatesFs embed.FS
	db          *pgxpool.Pool
}

func NewHandler(services *server.Service, logger *slog.Logger, conn *pgxpool.Pool) *ServerHandler {
	return &ServerHandler{
		services:    services,
		logger:      logger,
		TemplatesFs: templatesFs,
		db:          conn}
}

func (sh *ServerHandler) InitHandlers() http.Handler {
	mux := chi.NewMux()
	mux.Use(sh.RequestLogger, sh.WithCompressedResponse)

	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", sh.getMetricValue)
	mux.Get("/", sh.showMetrics)

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VALUE>
	// Content-Type: text/plain
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", sh.handleUploads)

	mux.Post("/update/", sh.handleJSONUploads)
	mux.Post("/updates/", sh.handleBatchUploads)
	mux.Post("/value/", sh.getJSONMetricValue)

	mux.Get("/ping", sh.handleDBPing)

	return mux
}
