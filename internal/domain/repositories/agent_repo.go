// Package repositories provides interfaces and implementations for data storage and retrieval related to metrics.
package repositories

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// MetricsCollectionRepository defines methods for storing and retrieving metrics collection.
// This interface allows implementations to provide various data storage backends.
type MetricsCollectionRepository interface {
	// Store saves the given gauge metrics into the repository.
	// It returns an error if the operation fails.
	Store(gaugeMetrics map[entities.MetricName]entities.Gauge) error
	
	// GetAll retrieves all metrics from the repository.
	// It returns a pointer to a MetricsCollection and an error if the operation fails.
	GetAll() (*entities.MetricsCollection, error)
}

// AgentRepository is a concrete implementation of the MetricsCollectionRepository interface.
// It acts as a bridge to the underlying metrics storage mechanism.
type AgentRepository struct {
	MetricsCollectionRepository
}

// NewAgentRepository creates a new instance of AgentRepository with the provided storage and logger.
// It initializes the repository with a metrics collection storage implementation.
func NewAgentRepository(store *storage.MetricsCollection, logger *slog.Logger) *AgentRepository {
	return &AgentRepository{
		MetricsCollectionRepository: NewMetricsCollectionMemRepository(store, logger),
	}
}
