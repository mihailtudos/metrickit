package service

import (
	"fmt"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"log/slog"
	"net/http"
	"runtime"
)

type MetricsCollectionService struct {
	mRepo  repositories.MetricsCollectionRepository
	logger *slog.Logger
}

func NewMetricsCollectionService(repo repositories.MetricsCollectionRepository, logger *slog.Logger) *MetricsCollectionService {
	return &MetricsCollectionService{mRepo: repo, logger: logger}
}

func (m *MetricsCollectionService) Collect() {
	m.logger.Info("collecting metrics")
	currMetric := runtime.MemStats{}
	runtime.ReadMemStats(&currMetric)

	m.mRepo.Store(&currMetric)
}

func (m *MetricsCollectionService) Send(serverAddr string) {
	metrics := m.mRepo.GetAll()
	// publish counter type metrics
	m.logger.Info("publishing counter metrics")
	for _, v := range metrics.CounterMetrics {
		err := publishMetric(serverAddr, v)
		if err != nil {
			m.logger.Error(fmt.Sprintf("something went wrong when publishing the counter metrics %s", err.Error()))
		}
	}

	// publish gauge type metrics
	m.logger.Info("publishing gauge metrics")
	for _, v := range metrics.GaugeMetrics {
		err := publishMetric(serverAddr, v)
		if err != nil {
			m.logger.Error(fmt.Sprintf("something went wrong when publishing the gauge metrics %s", err.Error()))
		}
	}
}

func (m *MetricsCollectionService) Clear() {
	m.logger.Info("clearing metrics")
	m.mRepo.Clear()
}

func publishMetric(ServerAddr string, metric any) error {
	url := ""

	switch v := metric.(type) {
	case entities.CounterMetric:
		url = fmt.Sprintf("http://%s/update/%s/%s/%v", ServerAddr, entities.CounterMetricName, v.Name, v.Value)
	case entities.GaugeMetric:
		url = fmt.Sprintf("http://%s/update/%s/%s/%v", ServerAddr, entities.GaugeMetricName, v.Name, v.Value)
	}

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
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
