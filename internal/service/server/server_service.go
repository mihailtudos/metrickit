package server

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type Metrics interface {
	Create(metric entities.Metrics) error
	Get(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error)
	GetAll() (*storage.MetricsStorage, error)
	GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)
}

type Service struct {
	MetricsService Metrics
}

func NewMetricsService(repository *repositories.Repository, logger *slog.Logger) *Service {
	return &Service{
		MetricsService: NewMetricService(repository.MetricsRepository, logger),
	}
}
