package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
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

	c.logger.Log(context.Background(), slog.LevelInfo, fmt.Sprintf("storing metric %s %v", key, v))
	err = c.cRepo.Create(key, entities.Counter(v))
	if err != nil {
		return fmt.Errorf("failed to create metric counter with key=%s val=%s due to: %w", key, val, err)
	}

	return nil
}

func (c *CounterMetricService) Get(key string) (entities.Counter, bool) {
	return c.cRepo.Get(key)
}

func (c *CounterMetricService) GetAll() map[string]entities.Counter {
	return c.cRepo.GetAll()
}
