package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type GaugeRepository interface {
	Create(key entities.MetricName, gauge entities.Gauge) error
	Get(key entities.MetricName) (entities.Gauge, error)
	GetAll() (map[entities.MetricName]entities.Gauge, error)
}

type CounterRepository interface {
	Create(key entities.MetricName, counter entities.Counter) error
	Get(key entities.MetricName) (entities.Counter, error)
	GetAll() (map[entities.MetricName]entities.Counter, error)
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
