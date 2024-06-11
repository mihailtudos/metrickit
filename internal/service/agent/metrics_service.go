package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/pkg/compressor"
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
	m.logger.DebugContext(context.Background(), "collecting metrics...")

	currMetric := runtime.MemStats{}
	runtime.ReadMemStats(&currMetric)

	if err := m.mRepo.Store(&currMetric); err != nil {
		return fmt.Errorf("failed to store the metrics: %w", err)
	}

	return nil
}

func (m *MetricsCollectionService) Send(serverAddr string) error {
	url := fmt.Sprintf("http://%s/updates/", serverAddr)

	metrics, err := m.mRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to send the metrics: %w", err)
	}

	allMetrics := make([]entities.Metrics, 0, len(metrics.CounterMetrics)+len(metrics.CounterMetrics))

	m.logger.DebugContext(context.Background(), "publishing counter metrics")
	for k, v := range metrics.CounterMetrics {
		val := int64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	m.logger.DebugContext(context.Background(), "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		val := float64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	err = m.publishMetric(url, "application/json", allMetrics)
	if err != nil {
		m.logger.DebugContext(context.Background(),
			"publishing the counter metrics failed: "+err.Error())
	}

	return nil
}

var ErrJSONMarshal = errors.New("failed to marshal to JSON")

func (m *MetricsCollectionService) publishMetric(url, contentType string, metrics []entities.Metrics) error {
	mJSONStruct, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed serialize the metrics: %w", ErrJSONMarshal)
	}

	gzipBuffer, err := compressor.Compress(mJSONStruct)
	if err != nil {
		return errors.New("failed to compress metrics: " + err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(gzipBuffer))
	if err != nil {
		return errors.New("failed to create HTTP request: " + err.Error())
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	res, err := client.Do(req)
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

	m.logger.DebugContext(context.Background(), "published successfully", slog.String("metric", string(mJSONStruct)))
	return nil
}
