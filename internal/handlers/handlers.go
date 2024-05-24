package handlers

import (
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/internal/middleware"
	"github.com/mihailtudos/metrickit/internal/service/server"

	"github.com/go-chi/chi/v5"
)

type ServerHandler struct {
	services *server.Service
	logger   *slog.Logger
}

func NewHandler(services *server.Service, logger *slog.Logger) *ServerHandler {
	return &ServerHandler{services: services, logger: logger}
}

func (h *ServerHandler) InitHandlers() http.Handler {
	mux := chi.NewMux()
	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", h.getMetricValue)
	mux.Get("/", h.showMetrics)

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VALUE>
	// Content-Type: text/plain
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", h.handleUploads)

	mux.Post("/update", h.handleJSONUploads)
	mux.Post("/value", h.getJSONMetricValue)

	return middleware.RequestLogger(mux, h.logger)
}
