package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"html/template"
	"net/http"
	"path"
)

const staticDir = "./static"

func (h *HandlerStr) showMetrics(w http.ResponseWriter, r *http.Request) {
	fileName := "index.html"
	tmpl, err := template.ParseFiles(string(http.Dir(path.Join(staticDir, fileName))))
	if err != nil {
		fmt.Println(err)
		return
	}

	gauges := h.services.GaugeService.GetAll()
	counters := h.services.CounterService.GetAll()
	var memStore = storage.NewMemStorage()

	memStore.Counter = counters
	memStore.Gauge = gauges

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, memStore); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlerStr) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	val, err := isMetricAvailable(metricType, metricName, h)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch v := val.(type) {
	case entities.Counter:
		_, _ = fmt.Fprintf(w, "%v", v)
	case entities.Gauge:
		_, _ = fmt.Fprintf(w, "%v", v)
	default:
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func isMetricAvailable(metricType, metricName string, h *HandlerStr) (any, error) {
	if metricType == entities.CounterMetricName {
		counterValue, ok := h.services.CounterService.Get(metricName)
		if ok {
			return counterValue, nil
		}
	}

	if metricType == entities.GaugeMetricName {
		gaugeValue, ok := h.services.GaugeService.Get(metricName)
		if ok {
			fmt.Println(gaugeValue)
			return gaugeValue, nil
		}
	}

	return nil, errors.New("not found")
}
