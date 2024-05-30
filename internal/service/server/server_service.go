package server

import (
	"errors"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

var ErrInvalidValue = errors.New("invalid value given")

type CounterServiceInterface interface {
	Create(metric entities.Metrics) error
	Get(key entities.MetricName) (entities.Counter, error)
	GetAll() (map[entities.MetricName]entities.Counter, error)
}

type GaugeServiceInterface interface {
	Create(metric entities.Metrics) error
	Get(key entities.MetricName) (entities.Gauge, error)
	GetAll() (map[entities.MetricName]entities.Gauge, error)
}

type Service struct {
	CounterService CounterServiceInterface
	GaugeService   GaugeServiceInterface
}

func NewService(repository *repositories.Repository, logger *slog.Logger) *Service {
	return &Service{
		GaugeService:   NewGaugeService(repository.GaugeRepository, logger),
		CounterService: NewCounterService(repository.CounterRepository, logger),
	}
}
