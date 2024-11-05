package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// formatBodyMessageErrors formats and returns an error message for body reading errors.
func formatBodyMessageErrors(err error) error {
	return fmt.Errorf("error reading body: %w", err)
}

// handleUploads handles metric uploads, validating and storing them based on their type.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a metric
// @Description Uploads a metric of type gauge or counter. Returns status OK if successful.
// @Tags metrics
// @Accept text/plain
// @Produce text/plain
// @Param metricType path string true "Type of metric (counter/gauge)"
// @Param metricName path string true "Name of the metric"
// @Param metricValue path string true "Value of the metric"
// @Success 200 {string} string "Metric uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/{metricType}/{metricName}/{metricValue} [post]
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

// handleBatchUploads handles batch metric uploads, validating the metrics and storing them.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a batch of metrics
// @Description Uploads multiple metrics in a single request. Returns status OK if successful.
// @Tags metrics
// @Accept application/json
// @Produce application/json
// @Param body []entities.Metrics true "List of metrics"
// @Success 200 {string} string "Metrics uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/batch [post]
func (sh *ServerHandler) handleBatchUploads(w http.ResponseWriter, r *http.Request) {
	metrics := make([]entities.Metrics, 0)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	if sh.secret != "" {
		if !isBodyValid(body, r.Header.Get("HashSHA256"), sh.secret) {
			sh.logger.DebugContext(r.Context(),
				"request body failed integrity check")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		sh.logger.DebugContext(r.Context(),
			"request body passed integrity check")
	}

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusInternalServerError)
		return
	}

	// Validate each metric type
	for _, metric := range metrics {
		if !isValidMetricType(metric.MType) {
			sh.logger.DebugContext(r.Context(),
				"unsupported metric type",
				slog.String("metric_type", metric.MType))
			http.Error(w, "Unsupported metric type", http.StatusInternalServerError)
			return
		}
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

// handleJSONUploads handles JSON metric uploads, validating and storing them.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a JSON metric
// @Description Uploads a single metric in JSON format. Returns status OK if successful.
// @Tags metrics
// @Accept application/json
// @Produce application/json
// @Param body entities.Metrics true "Metric"
// @Success 200 {object} entities.Metrics "Metric uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/json [post]
func (sh *ServerHandler) handleJSONUploads(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	// Handle JSON unmarshal error
	if err = json.Unmarshal(body, &metric); err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusBadRequest) // Changed to 400
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
		if err = sh.services.MetricsService.Create(metric); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to crete the counter metric",
				helpers.ErrAttr(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case entities.GaugeMetricName:
		if err = sh.services.MetricsService.Create(metric); err != nil {
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

// isMetricNameAndValuePresent checks if the metric has a valid name and value.
// It returns true if the metric is of type Counter and has a non-nil Delta and a non-empty ID,
// or if the metric is of type Gauge and has a non-nil Value and a non-empty ID.
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

// getResponseMetric retrieves the current value of the metric for response generation.
// If the metric is of type Gauge, it returns the metric as is.
// If the metric is of type Counter, it retrieves the current delta value from the MetricsService.
// It returns the metric and any error encountered during retrieval.
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

// isBodyValid verifies the integrity of the request body by comparing the provided hash with a computed hash.
// It returns true if the request hash is not empty and matches the computed hash using the provided secret.
func isBodyValid(data []byte, reqHash, secret string) bool {
	if reqHash == "" {
		return false
	}
	hashedStr := getHash(data, secret)
	return hashedStr == reqHash
}

// getHash computes the HMAC SHA-256 hash of the provided data using the provided secret.
// It returns the hex-encoded hash as a string.
func getHash(data []byte, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// isValidMetricType checks if the provided metric type is valid.
// It returns true if the metric type is either Counter or Gauge; otherwise, it returns false.
func isValidMetricType(mType string) bool {
	return mType == string(entities.CounterMetricName) || mType == string(entities.GaugeMetricName)
}
