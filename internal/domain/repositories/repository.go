package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type GaugeRepository interface {
	Create(key string, gauge entities.Gauge) error
}

type CounterRepository interface {
	Create(key string, counter entities.Counter) error
}

type Repository struct {
	GaugeRepository
	CounterRepository
}

func NewRepository(store *storage.MemStorage) *Repository {
	return &Repository{
		GaugeRepository:   NewGaugeMemRepository(store),
		CounterRepository: NewCounterMemRepository(store),
	}
}
