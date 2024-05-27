package handlers

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

var ErrUnknownMetric = errors.New("unknown metric type")
var ContentType = "Content-Type"

func (h *ServerHandler) showMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Println(h.TemplatesFs)
	tmpl, err := template.ParseFS(h.TemplatesFs, "templates/index.html")
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

	var memStore = storage.MemStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}

	memStore.Counter = counters
	memStore.Gauge = gauges

	w.Header().Set(ContentType, "text/html; charset=utf-8")
	err = tmpl.ExecuteTemplate(w, "index.html", memStore)

	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to execute template: "+err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *ServerHandler) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metric := entities.Metrics{ID: metricName, MType: metricType}
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

	w.Header().Set(ContentType, "text/plain; charset=utf-8")
	switch entities.MetricType(currentMetric.MType) {
	case entities.CounterMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Delta)
	case entities.GaugeMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Value)

	default:
		h.logger.ErrorContext(context.Background(), "failed identify the correct metric type")
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *ServerHandler) getJSONMetricValue(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set(ContentType, "application/json; charset=utf-8")
	jsonMetric, err := json.MarshalIndent(currentMetric, "", "  ")
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to marshal metric: "+err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(jsonMetric); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to write response: "+err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
