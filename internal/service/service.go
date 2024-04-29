package service

import (
	"errors"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"log/slog"
)

var ErrInvalidValue = errors.New("invalid value given")

type CounterService interface {
	Create(key string, val string) error
}

type GaugeService interface {
	Create(key string, val string) error
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
