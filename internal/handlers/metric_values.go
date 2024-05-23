package handlers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

const staticDir = "./static"

var ErrUnknownMetric = errors.New("unknown metric type")

func (h *HandlerStr) showMetrics(w http.ResponseWriter, r *http.Request) {
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

func (h *HandlerStr) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := entities.MetricType(chi.URLParam(r, "metricType"))
	metricName := entities.MetricName(chi.URLParam(r, "metricName"))

	val, err := isMetricAvailable(metricType, metricName, h)
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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch v := val.(type) {
	case entities.Counter:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", v)
	case entities.Gauge:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", v)
	default:
		h.logger.ErrorContext(context.Background(), "failed identify the correct metric type")
		w.WriteHeader(http.StatusNotFound)
	}
}

func isMetricAvailable(metricType entities.MetricType, metricName entities.MetricName, h *HandlerStr) (any, error) {
	if metricType == entities.CounterMetricName {
		counterValue, err := h.services.CounterService.Get(metricName)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("item was not found metric type %s and metric name %s: %w", metricType, metricName, err)
			}

			return nil, errors.New("failed to get the given metric: " + err.Error())
		}

		return counterValue, nil
	}

	if metricType == entities.GaugeMetricName {
		gaugeValue, err := h.services.GaugeService.Get(metricName)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("item not found with metric type %s and metric name %s: %w", metricType, metricName, err)
			}

			return nil, errors.New("failed to get the given metric: " + err.Error())
		}

		return gaugeValue, nil
	}

	return nil, ErrUnknownMetric
}
