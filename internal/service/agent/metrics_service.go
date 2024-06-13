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
	"github.com/mihailtudos/metrickit/pkg/helpers"
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
	ctx := context.Background()

	metrics, err := m.mRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to send the metrics: %w", err)
	}

	allMetrics := make([]entities.Metrics, 0, len(metrics.CounterMetrics)+len(metrics.CounterMetrics))

	m.logger.DebugContext(ctx, "publishing counter metrics")
	for k, v := range metrics.CounterMetrics {
		val := int64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	m.logger.DebugContext(ctx, "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		val := float64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	err = m.publishMetric(ctx, url, "application/json", allMetrics)
	if err != nil {
		m.logger.ErrorContext(ctx,
			"publishing the counter metrics failed: ",
			helpers.ErrAttr(err))
	}

	return nil
}

var ErrJSONMarshal = errors.New("failed to marshal to JSON")

func (m *MetricsCollectionService) publishMetric(ctx context.Context, url,
	contentType string, metrics []entities.Metrics) error {
	mJSONStruct, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed serialize the metrics: %w", ErrJSONMarshal)
	}

	gzipBuffer, err := compressor.Compress(mJSONStruct)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(gzipBuffer))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post metric: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			m.logger.ErrorContext(ctx, "failed to close the body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to publish the metric " + res.Status)
	}

	m.logger.DebugContext(ctx, "published successfully", slog.String("metric", string(mJSONStruct)))
	return nil
}
