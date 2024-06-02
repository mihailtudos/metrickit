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
	metrics, err := m.mRepo.GetAll()
	if err != nil {
		m.logger.ErrorContext(context.Background(), fmt.Sprintf("failed to send the metrics: %v", err))
		return errors.New("failed to send the metrics: " + err.Error())
	}

	m.logger.DebugContext(context.Background(), "publishing counter metrics")
	for k, v := range metrics.CounterMetrics {
		val := int64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: &val,
		}
		err := m.publishMetric(serverAddr, "application/json", &metric)
		if err != nil {
			m.logger.DebugContext(context.Background(),
				"publishing the counter metrics failed: "+err.Error())
		}
	}

	// publish gauge type metrics
	m.logger.DebugContext(context.Background(), "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		val := float64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: &val,
		}
		err := m.publishMetric(serverAddr, "application/json", &metric)
		if err != nil {
			m.logger.ErrorContext(context.Background(),
				"publishing the gauge metrics failed: "+err.Error())
		}
	}

	return nil
}

func (m *MetricsCollectionService) publishMetric(serverAddr, contentType string, metric *entities.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", serverAddr)

	mJSONStruct, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to public metric: %w", entities.ErrJSONMarshal)
	}

	gzipBuffer, err := compressor.Compress(mJSONStruct)
	if err != nil {
		return errors.New("failed to compress metric: " + err.Error())
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
