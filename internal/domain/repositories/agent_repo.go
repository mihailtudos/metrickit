package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"log/slog"
)

type MetricsCollectionRepository interface {
	Store(gaugeMetrics map[entities.MetricName]entities.Gauge) error
	GetAll() (*entities.MetricsCollection, error)
}

type AgentRepository struct {
	MetricsCollectionRepository
}

func NewAgentRepository(store *storage.MetricsCollection, logger *slog.Logger) *AgentRepository {
	return &AgentRepository{
		MetricsCollectionRepository: NewMetricsCollectionMemRepository(store, logger),
	}
}
