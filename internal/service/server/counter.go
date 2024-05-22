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

func (c *CounterMetricService) Create(key string, val string) error {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return ErrInvalidValue
	}

	c.logger.DebugContext(context.Background(), fmt.Sprintf("storing metric %s %v", key, v))
	err = c.cRepo.Create(key, entities.Counter(v))
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%s due to: %w", key, val, err)
	}

	return nil
}

func (c *CounterMetricService) Get(key string) (entities.Counter, error) {
	item, err := c.cRepo.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Counter(0), err
		}

		return entities.Counter(0), errors.New("failed to get the metric: " + err.Error())
	}

	return item, nil
}

func (c *CounterMetricService) GetAll() (map[string]entities.Counter, error) {
	items, err := c.cRepo.GetAll()
	if err != nil {
		return nil, errors.New("failed to get the counter metrics: " + err.Error())
	}

	return items, nil
}
