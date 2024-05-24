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

type GaugeMetricService struct {
	gRepo  repositories.GaugeRepository
	logger *slog.Logger
}

func NewGaugeService(repo repositories.GaugeRepository, logger *slog.Logger) *GaugeMetricService {
	return &GaugeMetricService{gRepo: repo, logger: logger}
}

func (g *GaugeMetricService) Create(metric entities.Metrics) error {
	g.logger.DebugContext(context.Background(),
		fmt.Sprintf("setting gauge: %s to %v",
			metric.ID, *metric.Value))
	err := g.gRepo.Create(metric)
	if err != nil {
		return fmt.Errorf("unable to create the gauge metric with key=%s and value=%v due to: %w",
			metric.ID, *metric.Value, err)
	}

	return nil
}

func (g *GaugeMetricService) Get(key entities.MetricName) (entities.Gauge, error) {
	item, err := g.gRepo.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Gauge(0), err
		}

		return entities.Gauge(0), errors.New("failed to get the metric: " + err.Error())
	}

	return item, nil
}

func (g *GaugeMetricService) GetAll() (map[entities.MetricName]entities.Gauge, error) {
	items, err := g.gRepo.GetAll()
	if err != nil {
		return nil, errors.New("failed to get the gauge metrics: " + err.Error())
	}

	return items, nil
}
