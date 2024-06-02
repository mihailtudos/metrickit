package handlers

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/internal/service/server"

	"github.com/go-chi/chi/v5"
)

//go:embed templates
var templatesFs embed.FS

type ServerHandler struct {
	services    *server.Service
	logger      *slog.Logger
	TemplatesFs embed.FS
}

func NewHandler(services *server.Service, logger *slog.Logger) *ServerHandler {
	return &ServerHandler{services: services, logger: logger, TemplatesFs: templatesFs}
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
	mux.Post("/value/", sh.getJSONMetricValue)

	return mux
}
