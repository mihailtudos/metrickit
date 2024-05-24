package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

type MetricsCollectionService struct {
	mRepo  repositories.MetricsCollectionRepository
	logger *slog.Logger
}

func NewMetricsCollectionService(repo repositories.MetricsCollectionRepository,
	logger *slog.Logger) *MetricsCollectionService {
	return &MetricsCollectionService{mRepo: repo, logger: logger}
}

func (m *MetricsCollectionService) Collect() error {
	currMetric := runtime.MemStats{}
	runtime.ReadMemStats(&currMetric)

	if err := m.mRepo.Store(&currMetric); err != nil {
		return errors.New("failed to store the metrics " + err.Error())
	}

	return nil
}

func (m *MetricsCollectionService) Send(serverAddr string) error {
	metrics, err := m.mRepo.GetAll()
	if err != nil {
		m.logger.ErrorContext(context.Background(), fmt.Sprintf("failed to send the metrics: %v", err))
		return errors.New("failed to send the metrics: " + err.Error())
	}

	m.logger.DebugContext(context.Background(), "publishing counter metrics")
	for k, v := range metrics.CounterMetrics {
		url := fmt.Sprintf("http://%s/update", serverAddr)
		val := int64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: &val,
		}
		err := m.publishMetric(url, metric)
		if err != nil {
			if errors.Is(err, entities.ErrJSONMarshal) {
				m.logger.DebugContext(context.Background(), "failed to marshal struct to JSON : "+err.Error())
			}
			m.logger.DebugContext(context.Background(),
				"something went wrong when publishing the counter metrics: "+err.Error())
		}
	}

	// publish gauge type metrics
	m.logger.DebugContext(context.Background(), "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		url := fmt.Sprintf("http://%s/update", serverAddr)
		val := float64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: &val,
		}
		err := m.publishMetric(url, metric)
		if err != nil {
			if errors.Is(err, entities.ErrJSONMarshal) {
				m.logger.DebugContext(context.Background(), "failed to marshal struct to JSON : "+err.Error())
			}
			m.logger.ErrorContext(context.Background(),
				"something went wrong when publishing the gauge metrics: "+err.Error())
		}
	}

	return nil
}

func (m *MetricsCollectionService) publishMetric(url string, metrics entities.Metrics) error {
	mJsonStruct, err := json.Marshal(metrics)
	if err != nil {
		return entities.ErrJSONMarshal
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(mJsonStruct))
	if err != nil {
		return errors.New("failed to post metric" + err.Error())
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			m.logger.ErrorContext(context.Background(), "failed to close the body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to publish the metric " + res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		m.logger.ErrorContext(context.Background(), "failed to read response body"+err.Error())
		return errors.New("failed to read response body " + err.Error())
	}

	m.logger.DebugContext(context.Background(), "published successfully: ", slog.String("response", string(body)))
	return nil
}
