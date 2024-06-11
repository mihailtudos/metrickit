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
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

const errorMessageFormat = "error reading body: %s"

func (sh *ServerHandler) handleUploads(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// return http.StatusNotFound if metric type is not provided
	if entities.MetricType(metricType) != entities.GaugeMetricName &&
		entities.MetricType(metricType) != entities.CounterMetricName ||
		metricValue == "" {
		sh.logger.DebugContext(r.Context(),
			"invalid metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch entities.MetricType(metricType) {
	case entities.CounterMetricName:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to convert counter value to integer",
				slog.String("delta", metricValue))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		metric := entities.Metrics{ID: metricName, MType: metricType, Delta: &delta}
		if err = sh.services.MetricsService.Create(metric); err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to create the metric",
				helpers.ErrAttr(err),
				slog.String("url", r.RequestURI))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to convert gauge value to float64",
				slog.String("value", metricValue))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		metric := entities.Metrics{ID: metricName, MType: metricType, Value: &value}
		if err = sh.services.MetricsService.Create(metric); err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to create the metric",
				helpers.ErrAttr(err),
				slog.String("url", r.RequestURI))
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

func (sh *ServerHandler) handleBatchUploads(w http.ResponseWriter, r *http.Request) {
	metrics := make([]entities.Metrics, 0)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, fmt.Sprintf(errorMessageFormat, err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, fmt.Sprintf(errorMessageFormat, err), http.StatusBadRequest)
		return
	}

	sh.logger.DebugContext(r.Context(),
		fmt.Sprintf("received batch of %d metrics", len(metrics)))
	// return http.StatusNotFound if metric type is not provided

	w.Header().Set("Content-Type", "application/json")
	err = sh.services.MetricsService.StoreMetricsBatch(metrics)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to batch write the metrics",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (sh *ServerHandler) handleJSONUploads(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, fmt.Sprintf(errorMessageFormat, err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	err = json.Unmarshal(body, &metric)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, fmt.Sprintf(errorMessageFormat, err), http.StatusBadRequest)
		return
	}

	sh.logger.DebugContext(r.Context(), "received", slog.String("metric", string(body)))
	// return http.StatusNotFound if metric type is not provided
	if metric.MType != string(entities.GaugeMetricName) &&
		metric.MType != string(entities.CounterMetricName) {
		sh.logger.DebugContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if !isMetricNameAndValuePresent(metric) {
		sh.logger.DebugContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch entities.MetricType(metric.MType) {
	case entities.CounterMetricName:
		if err := sh.services.MetricsService.Create(metric); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to crete the counter metric",
				helpers.ErrAttr(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case entities.GaugeMetricName:
		if err := sh.services.MetricsService.Create(metric); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to create the gauge metric",
				helpers.ErrAttr(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	default:
		sh.logger.DebugContext(r.Context(),
			"no valid metric received")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	updatedMetric, err := sh.getResponseMetric(metric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed generate response",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(updatedMetric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal the response",
			helpers.ErrAttr(err),
			slog.String(bodyKey, string(body)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to write the response",
			helpers.ErrAttr(err),
			slog.String("response", string(response)))
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

func (sh *ServerHandler) getResponseMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		return &metric, nil
	} else {
		currentDelta, err := sh.services.MetricsService.Get(entities.MetricName(metric.ID), entities.CounterMetricName)
		if err != nil {
			return nil, fmt.Errorf("failed to get the counter value %w", err)
		}

		return &currentDelta, nil
	}
}
