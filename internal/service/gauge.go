package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

type GaugeMetricService struct {
	gRepo  repositories.GaugeRepository
	logger *slog.Logger
}

func NewGaugeService(repo repositories.GaugeRepository, logger *slog.Logger) *GaugeMetricService {
	return &GaugeMetricService{gRepo: repo, logger: logger}
}

func (g *GaugeMetricService) Create(key string, val string) error {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return ErrInvalidValue
	}

	g.logger.Log(context.Background(), slog.LevelInfo, fmt.Sprintf("setting gauge: %s to %v", key, v))
	err = g.gRepo.Create(key, entities.Gauge(v))
	if err != nil {
		return fmt.Errorf("unable to create the gauge metric with key=%s and value=%s due to: %w", key, val, err)
	}

	return nil
}

func (g *GaugeMetricService) Get(key string) (entities.Gauge, bool) {
	return g.gRepo.Get(key)
}

func (g *GaugeMetricService) GetAll() map[string]entities.Gauge {
	return g.gRepo.GetAll()
}
