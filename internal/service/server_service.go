package service

import (
	"errors"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

var ErrInvalidValue = errors.New("invalid value given")

type CounterService interface {
	Create(key string, val string) error
	Get(key string) (entities.Counter, bool)
	GetAll() map[string]entities.Counter
}

type GaugeService interface {
	Create(key string, val string) error
	Get(key string) (entities.Gauge, bool)
	GetAll() map[string]entities.Gauge
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
