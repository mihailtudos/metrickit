package handlers

import (
	"github.com/mihailtudos/metrickit/internal/service"
	"log/slog"
	"net/http"
)

type Handler struct {
	services *service.Service
	logger   *slog.Logger
}

func NewHandler(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{services: services, logger: logger}
}

func (h *Handler) InitHandlers() http.Handler {
	mux := http.NewServeMux()

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.HandleFunc("/update/counter/", h.handleUploads)
	mux.HandleFunc("/update/gauge/", h.handleUploads)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return mux
}
