package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"

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
	m.logger.DebugContext(context.Background(), "collecting metrics...")

	currMetric := runtime.MemStats{}
	runtime.ReadMemStats(&currMetric)

	if err := m.mRepo.Store(&currMetric); err != nil {
		return errors.New("failed to send the metrics " + err.Error())
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
		val := strconv.Itoa(int(v))
		url := fmt.Sprintf("http://%s/update/%s/%s/%v", serverAddr, entities.CounterMetricName, k, val)
		err := m.publishMetric(url)
		if err != nil {
			m.logger.DebugContext(context.Background(),
				"something went wrong when publishing the counter metrics: "+err.Error())
		}
	}

	// publish gauge type metrics
	m.logger.DebugContext(context.Background(), "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		val := strconv.FormatFloat(float64(v), 'f', -1, 64)
		url := fmt.Sprintf("http://%s/update/%s/%s/%v", serverAddr, entities.GaugeMetricName, k, val)
		err := m.publishMetric(url)
		if err != nil {
			m.logger.ErrorContext(context.Background(),
				"something went wrong when publishing the gauge metrics: "+err.Error())
		}
	}

	return nil
}

func (m *MetricsCollectionService) publishMetric(url string) error {
	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return errors.New("failed to post metric" + err.Error())
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			m.logger.ErrorContext(context.Background(), "failed to close the body")
		}
	}()

	if res.StatusCode == http.StatusOK {
		m.logger.InfoContext(context.Background(), "published successfully: ", slog.String("url", url))
		return nil
	}

	return errors.New("failed to publish the metric " + res.Status)
}
