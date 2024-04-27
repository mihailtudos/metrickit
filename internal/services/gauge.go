package services

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
		return err
	}

	g.logger.Info(fmt.Sprintf("setting gauge: %s to %v", key, v))
	return g.gRepo.Create(key, entities.Gauge(v))
}
