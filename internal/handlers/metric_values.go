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
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

var ErrUnknownMetric = errors.New("unknown metric type")

const bodyKey = "body"

func (sh *ServerHandler) showMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(sh.TemplatesFs, "templates/index.html")
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to parse the template: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	metrics, err := sh.services.MetricsService.GetAll()
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to get the metrics: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "text/html; charset=utf-8")
	err = tmpl.ExecuteTemplate(w, "index.html", metrics)

	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to execute template: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (sh *ServerHandler) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metric := entities.Metrics{ID: metricName, MType: metricType}
	currentMetric, err := sh.getMetric(metric)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to get the metric struct",
			helpers.ErrAttr(err))
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUnknownMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sh.logger.ErrorContext(r.Context(),
			"failed to get metric: ",
			helpers.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "text/plain; charset=utf-8")
	switch entities.MetricType(currentMetric.MType) {
	case entities.CounterMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Delta)
	case entities.GaugeMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Value)

	default:
		sh.logger.ErrorContext(r.Context(),
			"failed identify the correct metric type")
		w.WriteHeader(http.StatusNotFound)
	}
}

func (sh *ServerHandler) getJSONMetricValue(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to read body: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal request body: ",
			helpers.ErrAttr(err),
			slog.String(bodyKey, string(body)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	currentMetric, err := sh.getMetric(metric)
	if err != nil {
		sh.logger.DebugContext(context.Background(),
			"failed to get the metric struct",
			helpers.ErrAttr(err))
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUnknownMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sh.logger.ErrorContext(context.Background(),
			"failed to get metric: ",
			helpers.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "application/json; charset=utf-8")
	jsonMetric, err := json.MarshalIndent(currentMetric, "", "  ")
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal metric: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(jsonMetric); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to write response: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (sh *ServerHandler) getMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.CounterMetricName {
		record, err := sh.services.MetricsService.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, fmt.Errorf("failed to get the given metric: %w", err)
		}

		return &record, nil
	}

	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		record, err := sh.services.MetricsService.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, fmt.Errorf("failed to get the given metric: %w", err)
		}

		return &record, nil
	}

	return nil, ErrUnknownMetric
}

func (sh *ServerHandler) handleDBPing(w http.ResponseWriter, r *http.Request) {
	if err := sh.db.Ping(r.Context()); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to ping the DB",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}
