package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

func (h *ServerHandler) handleUploads(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// return http.StatusNotFound if metric type is not provided
	if entities.MetricType(metricType) != entities.GaugeMetricName &&
		entities.MetricType(metricType) != entities.CounterMetricName ||
		metricValue == "" {
		h.logger.ErrorContext(r.Context(), "invalid metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch entities.MetricType(metricType) {
	case entities.CounterMetricName:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "failed to convert counter value to integer", slog.String("delta", metricValue))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		metric := entities.Metrics{ID: metricName, MType: metricType, Delta: &delta}
		if err = h.services.CounterService.Create(metric); err != nil {
			h.logger.ErrorContext(r.Context(), "failed to create the metric", slog.String("err", err.Error()), slog.String("url", r.RequestURI))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "failed to convert gauge value to float64", slog.String("value", metricValue))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		metric := entities.Metrics{ID: metricName, MType: metricType, Value: &value}
		if err = h.services.GaugeService.Create(metric); err != nil {
			h.logger.ErrorContext(r.Context(), "failed to create the metric", slog.String("err", err.Error()), slog.String("url", r.RequestURI))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Upon successful reception, return StatusOK
	w.WriteHeader(http.StatusOK)
}

func (h *ServerHandler) handleJSONUploads(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to read request body")
		http.Error(w, fmt.Sprintf("error reading body: %s", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.ErrorContext(r.Context(), "failed to close request body")
		}
	}()

	err = json.Unmarshal(body, &metric)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to unmarshal the request", slog.String("body", string(body)))
		http.Error(w, fmt.Sprintf("error reading body: %s", err), http.StatusBadRequest)
		return
	}

	h.logger.DebugContext(r.Context(), "received", slog.String("metric", string(body)))
	// return http.StatusNotFound if metric type is not provided
	if metric.MType != string(entities.GaugeMetricName) &&
		metric.MType != string(entities.CounterMetricName) {
		h.logger.ErrorContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if !isMetricNameAndValuePresent(metric) {
		h.logger.ErrorContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch entities.MetricType(metric.MType) {
	case entities.CounterMetricName:
		if err := h.services.CounterService.Create(metric); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case entities.GaugeMetricName:
		if err := h.services.GaugeService.Create(metric); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	updatedMetric, err := h.getResponseMetric(metric)
	response, err := json.Marshal(updatedMetric)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to marshal the response", slog.String("body", string(body)))
		http.Error(w, fmt.Sprintf("failed to marshal the response: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to write the response", slog.String("response", string(response)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func isMetricNameAndValuePresent(metric entities.Metrics) bool {
	if metric.MType == string(entities.CounterMetricName) &&
		metric.Delta != nil &&
		metric.ID != "" {
		return true
	}

	if metric.MType == string(entities.GaugeMetricName) &&
		metric.Value != nil &&
		metric.ID != "" {
		return true
	}

	return false
}

func (h *ServerHandler) getResponseMetric(metric entities.Metrics) (*entities.Metrics, error) {
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
