package service

import (
	"fmt"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"log/slog"
	"strconv"
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

	g.logger.Info(fmt.Sprintf("setting gauge: %s to %v", key, v))
	return g.gRepo.Create(key, entities.Gauge(v))
}

func (g *GaugeMetricService) Get(key string) (entities.Gauge, bool) {
	return g.gRepo.Get(key)
}

func (g *GaugeMetricService) GetAll() map[string]entities.Gauge {
	return g.gRepo.GetAll()
}
