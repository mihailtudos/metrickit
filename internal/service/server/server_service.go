package server

import (
	"errors"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

var ErrInvalidValue = errors.New("invalid value given")

type CounterService interface {
	Create(key entities.MetricName, val string) error
	Get(key entities.MetricName) (entities.Counter, error)
	GetAll() (map[entities.MetricName]entities.Counter, error)
}

type GaugeService interface {
	Create(key entities.MetricName, val string) error
	Get(key entities.MetricName) (entities.Gauge, error)
	GetAll() (map[entities.MetricName]entities.Gauge, error)
}

type Service struct {
	CounterService
	GaugeService
}

func NewService(repository *repositories.Repository, logger *slog.Logger) *Service {
	return &Service{
		GaugeService:   NewGaugeService(repository.GaugeRepository, logger),
		CounterService: NewCounterService(repository.CounterRepository, logger),
	}
}
