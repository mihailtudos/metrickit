package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

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
	v := strconv.Itoa(int(*metric.Delta))

	c.logger.DebugContext(context.Background(), fmt.Sprintf("storing metric %s %v", metric.ID, v))
	err := c.cRepo.Create(entities.MetricName(metric.ID), entities.Counter(*metric.Delta))
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%s due to: %w", metric.ID, v, err)
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
