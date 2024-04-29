package service

import (
	"fmt"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"log/slog"
	"strconv"
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

	c.logger.Info(fmt.Sprintf("storing metric %s %v", key, v))
	return c.cRepo.Create(key, entities.Counter(v))
}
