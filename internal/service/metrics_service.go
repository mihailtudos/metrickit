package service

import (
	"context"
	"fmt"
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

func (m *MetricsCollectionService) Collect() {
	m.logger.Log(context.Background(), slog.LevelInfo, "collecting metrics")

	currMetric := runtime.MemStats{}
	runtime.ReadMemStats(&currMetric)

	m.mRepo.Store(&currMetric)
}

func (m *MetricsCollectionService) Send(serverAddr string) {
	metrics := m.mRepo.GetAll()
	// publish counter type metrics
	m.logger.Log(context.Background(), slog.LevelInfo, "publishing counter metrics")

	for _, v := range metrics.CounterMetrics {
		err := publishMetric(serverAddr, v)
		if err != nil {
			m.logger.Log(context.Background(),
				slog.LevelError,
				"something went wrong when publishing the counter metrics: "+err.Error())
		}
	}

	// publish gauge type metrics
	m.logger.Log(context.Background(), slog.LevelInfo, "publishing gauge metrics")
	for _, v := range metrics.GaugeMetrics {
		err := publishMetric(serverAddr, v)
		if err != nil {
			m.logger.Log(context.Background(),
				slog.LevelError,
				"something went wrong when publishing the gauge metrics: "+err.Error())
		}
	}
}

func (m *MetricsCollectionService) Clear() {
	m.logger.Log(context.Background(), slog.LevelInfo, "clearing metrics")
	m.mRepo.Clear()
}

func publishMetric(serverAddr string, metric any) error {
	url := ""

	switch v := metric.(type) {
	case entities.CounterMetric:
		url = fmt.Sprintf("http://%s/update/%s/%s/%v", serverAddr, entities.CounterMetricName, v.Name, v.Value)
	case entities.GaugeMetric:
		url = fmt.Sprintf("http://%s/update/%s/%s/%v", serverAddr, entities.GaugeMetricName, v.Name, v.Value)
	}

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("unable to post the metric %s due to %w", url, err)
	}

	defer func(res *http.Response) {
		if err := res.Body.Close(); err != nil {
			fmt.Println("failed to close the body")
		}
	}(res)

	if res.StatusCode == http.StatusOK {
		fmt.Println("published successfully: ", url)
	}

	return nil
}
