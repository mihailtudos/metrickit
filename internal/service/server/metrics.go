// Package server provides the MetricsService, which offers methods for
// creating, retrieving, and managing metrics. It interacts with a repository
// to store and fetch metrics data, and utilizes a logger for debugging and error tracking.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// MetricsService is responsible for managing metrics. It interacts with
// a repository for data storage and retrieval, and uses a logger for
// logging purposes.
type MetricsService struct {
	repo   repositories.MetricsRepository // Repository for metric storage and retrieval
	logger *slog.Logger                   // Logger for debug and error messages
}

// NewMetricService creates a new MetricsService instance with the
// specified repository and logger.
func NewMetricService(repo repositories.MetricsRepository, logger *slog.Logger) *MetricsService {
	return &MetricsService{repo: repo, logger: logger}
}

// Create adds a new metric to the repository. It logs the action and
// returns an error if the operation fails.
func (ms *MetricsService) Create(metric entities.Metrics) error {
	if metric.MType != string(entities.CounterMetricName) && metric.MType != string(entities.GaugeMetricName) {
		return fmt.Errorf("metric service: invalid metric type: %s", metric.MType)
	}

	ms.logger.DebugContext(context.Background(), fmt.Sprintf("updating %s metric", metric.ID))
	err := ms.repo.Create(metric)
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%v due to: %w", metric.ID, *metric.Delta, err)
	}

	return nil
}

// Get retrieves a specific metric by its key and type. It returns
// an error if the metric is not found or if an error occurs during retrieval.
func (ms *MetricsService) Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	item, err := ms.repo.Get(key, mType)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Metrics{}, fmt.Errorf("metric service: %w", err)
		}

		return entities.Metrics{}, fmt.Errorf("metric service: %w", err)
	}

	return item, nil
}

// GetAll retrieves all metrics from the repository. It returns
// an error if the retrieval fails.
func (ms *MetricsService) GetAll() (*storage.MetricsStorage, error) {
	items, err := ms.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the counter metrics: %w", err)
	}

	return items, nil
}

// GetAllByType retrieves all metrics of a specific type from the repository.
// It returns an error if the retrieval fails.
func (ms *MetricsService) GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	metrics, err := ms.repo.GetAllByType(mType)
	if err != nil {
		return nil, fmt.Errorf("metrics service: %w", err)
	}

	return metrics, nil
}

// StoreMetricsBatch stores a batch of metrics in the repository.
// It returns an error if the storage operation fails.
func (ms *MetricsService) StoreMetricsBatch(metrics []entities.Metrics) error {
	err := ms.repo.StoreMetricsBatch(metrics)
	if err != nil {
		return fmt.Errorf("metrics service %w", err)
	}

	return nil
}
