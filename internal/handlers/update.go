package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

func (h *ServerHandler) handleUploads(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to read request body")
		http.Error(w, fmt.Sprintf("error reading body: %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &metric)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to unmarshal the request", slog.String("body", string(body)))
		http.Error(w, fmt.Sprintf("error reading body: %s", err), http.StatusBadRequest)
		return
	}

	h.logger.DebugContext(r.Context(), "received new metric", slog.String("body", string(body)))

	// return http.StatusNotFound if metric type is not provided
	if metric.MType != string(entities.GaugeMetricName) && metric.MType != string(entities.CounterMetricName) {
		fmt.Println("invalid metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return http.StatusBadRequest if metric name or value submitted is not provided
	if !isMetricNameAndValuePresent(metric) {
		h.logger.ErrorContext(r.Context(),
			"invalid metric type or value",
			slog.String("metric", metric.MType),
			slog.Int64("delta", *metric.Delta),
			slog.Float64("value", *metric.Value),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch entities.MetricType(metric.MType) {
	case entities.CounterMetricName:
		if err := h.services.CounterService.Create(metric); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		if err := h.services.GaugeService.Create(metric); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
	updatedMetric, err := h.getUpdatedResponseMetric(metric)
	response, err := json.Marshal(updatedMetric)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to marshal the response", slog.String("body", string(body)))
		http.Error(w, fmt.Sprintf("failed to marshal the response: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func isMetricNameAndValuePresent(metric entities.Metrics) bool {
	if metric.MType == string(entities.CounterMetricName) && metric.Delta != nil {
		return true
	}

	if metric.MType == string(entities.GaugeMetricName) && metric.Value != nil {
		return true
	}

	return false
}

func (h *ServerHandler) getUpdatedResponseMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		return &metric, nil
	} else {
		currentDelta, err := h.services.CounterService.Get(entities.MetricName(metric.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to get the counter value %w", err)
		}

		intVal := int64(currentDelta)
		return &entities.Metrics{ID: metric.ID, MType: metric.MType, Delta: &intVal}, nil
	}
}
