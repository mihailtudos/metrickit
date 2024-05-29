package server

import (
	"errors"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

var ErrInvalidValue = errors.New("invalid value given")

// TODO(SSH): it's not a very good idea to declare your structure like that: it complexifies things
// and does not bring much to the table
// you should declare stuctures as they are and declare the "corresponding" interfaces where they are used
type CounterService interface {
	Create(metric entities.Metrics) error
	Get(key entities.MetricName) (entities.Counter, error)
	GetAll() (map[entities.MetricName]entities.Counter, error)
}

type GaugeService interface {
	Create(metric entities.Metrics) error
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
