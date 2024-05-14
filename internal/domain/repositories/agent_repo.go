package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"runtime"
)

type MetricsCollectionRepository interface {
	Store(stats *runtime.MemStats)
	GetAll() *entities.MetricsCollection
	Clear()
}

type AgentRepository struct {
	MetricsCollectionRepository
}

func NewAgentRepository(store *storage.MetricsCollection) *AgentRepository {
	return &AgentRepository{
		MetricsCollectionRepository: NewMetricsCollectionMemRepository(store),
	}
}
