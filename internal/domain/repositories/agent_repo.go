package repositories

import (
	"runtime"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type MetricsCollectionRepository interface {
	Store(stats *runtime.MemStats) error
	GetAll() (*entities.MetricsCollection, error)
}

type AgentRepository struct {
	MetricsCollectionRepository
}

func NewAgentRepository(store *storage.MetricsCollection) *AgentRepository {
	return &AgentRepository{
		MetricsCollectionRepository: NewMetricsCollectionMemRepository(store),
	}
}
