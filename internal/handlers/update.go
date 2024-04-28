package handlers

import (
	"fmt"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"net/http"
	"strings"
)

const (
	reqPartsLength = 5
	idxMetricType  = 2
	idxMetricName  = 3
	idxMetricVal   = 4
)

func (h *Handler) handleUploads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlParts := strings.Split(r.URL.Path, "/")

	// return http.StatusNotFound if metric type is not provided
	if len(urlParts) < reqPartsLength || !isMetricTypePresent(urlParts) {
		fmt.Println("missing metric type")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// return http.StatusBadRequest if metric name or value submitted is not provided
	if len(urlParts) < reqPartsLength || isMetricNameAndValueMissing(urlParts) {
		fmt.Println("missing metric name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch urlParts[idxMetricType] {
	case entities.CounterMetricName:
		if err := h.services.CounterService.Create(urlParts[idxMetricName], urlParts[idxMetricVal]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		if err := h.services.GaugeService.Create(urlParts[idxMetricName], urlParts[idxMetricVal]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}

	// Upon successful reception, return StatusOK
	w.WriteHeader(http.StatusOK)
}

func isMetricNameAndValueMissing(urlParts []string) bool {
	if isMetricNamePresent(urlParts) && isMetricValuePresent(urlParts) {
		return false
	}

	return true
}

func isMetricNamePresent(urlParts []string) bool {
	return urlParts[idxMetricName] != ""
}

func isMetricValuePresent(urlParts []string) bool {
	return urlParts[idxMetricVal] != ""
}

func isMetricTypePresent(urlParts []string) bool {
	return urlParts[idxMetricType] != ""
}
