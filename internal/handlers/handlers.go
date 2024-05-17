package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mihailtudos/metrickit/internal/service"
	"github.com/mihailtudos/metrickit/pkg/middlewars"
)

type ServerHandlers struct {
	services *service.Service
	logger   *slog.Logger
}

func NewHandler(services *service.Service, logger *slog.Logger) *ServerHandlers {
	return &ServerHandlers{services: services, logger: logger}
}

func (h *ServerHandlers) InitHandlers() http.Handler {
	mux := chi.NewMux()

	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", h.getMetricValue)
	mux.Get("/", h.showMetrics)

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VALUE>
	// Content-Type: text/plain
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", h.handleUploads)

	return middlewars.WithLogging(mux, h.logger)
}
