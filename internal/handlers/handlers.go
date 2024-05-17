package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/service"
)

type HandlerStr struct {
	services *service.Service
	logger   *slog.Logger
}

func NewHandler(services *service.Service, logger *slog.Logger) *HandlerStr {
	return &HandlerStr{services: services, logger: logger}
}

func (h *HandlerStr) InitHandlers() http.Handler {
	mux := chi.NewMux()

	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", h.getMetricValue)
	mux.Get("/", h.showMetrics)

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VALUE>
	// Content-Type: text/plain
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", h.handleUploads)

	return mux
}
