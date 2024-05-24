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

type CounterMetricService struct {
	cRepo  repositories.CounterRepository
	logger *slog.Logger
}

func NewCounterService(repo repositories.CounterRepository, logger *slog.Logger) *CounterMetricService {
	return &CounterMetricService{cRepo: repo, logger: logger}
}

func (c *CounterMetricService) Create(metric entities.Metrics) error {
	c.logger.DebugContext(context.Background(), fmt.Sprintf("updating %s metric", metric.ID))
	err := c.cRepo.Create(metric)
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%v due to: %w", metric.ID, *metric.Delta, err)
	}

	return nil
}

func (c *CounterMetricService) Get(key entities.MetricName) (entities.Counter, error) {
	item, err := c.cRepo.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Counter(0), err
		}

		return entities.Counter(0), errors.New("failed to get the metric: " + err.Error())
	}

	return item, nil
}

func (c *CounterMetricService) GetAll() (map[entities.MetricName]entities.Counter, error) {
	items, err := c.cRepo.GetAll()
	if err != nil {
		return nil, errors.New("failed to get the counter metrics: " + err.Error())
	}

	return items, nil
}
