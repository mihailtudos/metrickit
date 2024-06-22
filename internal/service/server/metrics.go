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

type MetricsService struct {
	repo   repositories.MetricsRepository
	logger *slog.Logger
}

func NewMetricService(repo repositories.MetricsRepository, logger *slog.Logger) *MetricsService {
	return &MetricsService{repo: repo, logger: logger}
}

func (ms *MetricsService) Create(metric entities.Metrics) error {
	ms.logger.DebugContext(context.Background(), fmt.Sprintf("updating %s metric", metric.ID))
	err := ms.repo.Create(metric)
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%v due to: %w", metric.ID, *metric.Delta, err)
	}

	return nil
}

func (ms *MetricsService) Get(key entities.MetricName,
	mType entities.MetricType) (entities.Metrics, error) {
	item, err := ms.repo.Get(key, mType)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Metrics{}, fmt.Errorf("metric service: %w", err)
		}

		return entities.Metrics{}, fmt.Errorf("metric service: %w", err)
	}

	return item, nil
}

func (ms *MetricsService) GetAll() (*storage.MetricsStorage, error) {
	items, err := ms.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the counter metrics: %w", err)
	}

	return items, nil
}

func (ms *MetricsService) GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	metrics, err := ms.repo.GetAllByType(mType)
	if err != nil {
		return nil, fmt.Errorf("metrics service: %w", err)
	}

	return metrics, nil
}

func (ms *MetricsService) StoreMetricsBatch(metrics []entities.Metrics) error {
	err := ms.repo.StoreMetricsBatch(metrics)
	if err != nil {
		return fmt.Errorf("metrics service %w", err)
	}

	return nil
}
