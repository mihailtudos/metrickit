package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

func (h *HandlerStr) handleUploads(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// return http.StatusNotFound if metric type is not provided
	if metricType != entities.GaugeMetricName && metricType != entities.CounterMetricName {
		fmt.Println("invalid metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return http.StatusBadRequest if metric name or value submitted is not provided
	if isMetricNameAndValueMissing(metricName, metricValue) {
		fmt.Println("missing metric name or value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch metricType {
	case entities.CounterMetricName:
		if err := h.services.CounterService.Create(metricName, metricValue); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		if err := h.services.GaugeService.Create(metricName, metricValue); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}

	// Upon successful reception, return StatusOK
	w.WriteHeader(http.StatusOK)
}

func isMetricNameAndValueMissing(metricName, metricValue string) bool {
	return metricName == "" || metricValue == ""
}
