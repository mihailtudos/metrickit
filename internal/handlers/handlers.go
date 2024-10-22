package handlers

import (
	"embed"
	"log/slog"
	"net/http"
	"net/http/pprof"

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
	secret      string
}

func NewHandler(services *server.Service, logger *slog.Logger,
	conn *pgxpool.Pool, secret string) http.Handler {
	handlers := &ServerHandler{
		services:    services,
		logger:      logger,
		TemplatesFs: templatesFs,
		db:          conn,
		secret:      secret}

	return handlers.registerRoutes()
}

func (sh *ServerHandler) registerRoutes() http.Handler {
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

	// Register pprof handlers
	mux.Get("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Get("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	mux.Handle("/debug/pprof/{profile}", http.HandlerFunc(pprof.Index))

	return mux
}
