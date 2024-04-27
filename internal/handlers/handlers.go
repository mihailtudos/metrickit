package handlers

import (
	"github.com/mihailtudos/metrickit/internal/services"
	"log/slog"
	"net/http"
)

type Handler struct {
	services *services.Service
	logger   *slog.Logger
}

func NewHandler(services *services.Service, logger *slog.Logger) *Handler {
	return &Handler{services: services, logger: logger}
}

func (h *Handler) InitHandlers() http.Handler {
	mux := http.NewServeMux()

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.HandleFunc("/update/counter/", h.handleUploads)
	mux.HandleFunc("/update/gauge/", h.handleUploads)

	mux.Handle("/", http.NotFoundHandler())
	return mux
}
