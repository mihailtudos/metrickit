package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type MetricsRepository interface {
	Create(metric entities.Metrics) error
	Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error)
	GetAll() (*storage.MetricsStorage, error)
	GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)
}

type Repository struct {
	MetricsRepository MetricsRepository
}

func NewRepository(store storage.Storage) *Repository {
	return &Repository{
		MetricsRepository: NewMetricsMemRepository(store),
	}
}
