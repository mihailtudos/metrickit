package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

const staticDir = "./static"

var ErrUnknownMetric = errors.New("unknown metric type")

func (h *ServerHandler) showMetrics(w http.ResponseWriter, r *http.Request) {
	fileName := "index.html"
	tmpl, err := template.ParseFiles(string(http.Dir(path.Join(staticDir, fileName))))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to parse the template: "+err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gauges, err := h.services.GaugeService.GetAll()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get the gauge metrics: "+err.Error())
		gauges = map[entities.MetricName]entities.Gauge{}
	}

	counters, err := h.services.CounterService.GetAll()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get the counter metrics: "+err.Error())
		counters = map[entities.MetricName]entities.Counter{}
	}

	var memStore = storage.NewMemStorage()

	memStore.Counter = counters
	memStore.Gauge = gauges

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, memStore)

	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to execute template: "+err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *ServerHandler) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to read body: "+err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to marshal request body: "+err.Error(), slog.String("body", string(body)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	currentMetric, err := h.getMetric(metric)
	if err != nil {
		h.logger.DebugContext(context.Background(), err.Error())
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUnknownMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.logger.ErrorContext(context.Background(), "failed to get metric: "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonMetric, err := json.MarshalIndent(currentMetric, "", "  ")
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to marshal metric: "+err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonMetric)
}

func (h *ServerHandler) getMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.CounterMetricName {
		counterValue, err := h.services.CounterService.Get(entities.MetricName(metric.ID))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, errors.New("failed to get the given metric: " + err.Error())
		}

		int64Val := int64(counterValue)
		return &entities.Metrics{ID: metric.ID, MType: metric.MType, Delta: &int64Val}, nil
	}

	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		gaugeValue, err := h.services.GaugeService.Get(entities.MetricName(metric.ID))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, errors.New("failed to get the given metric: " + err.Error())
		}

		float64Val := float64(gaugeValue)
		return &entities.Metrics{ID: metric.ID, MType: metric.MType, Value: &float64Val}, nil
	}

	return nil, ErrUnknownMetric
}
