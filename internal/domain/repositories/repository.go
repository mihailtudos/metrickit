// Package repositories provides interfaces and implementations for managing metric data storage and retrieval.
package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// MetricsRepository defines the methods for interacting with metric records.
// Implementations of this interface should provide concrete storage and retrieval mechanisms.
type MetricsRepository interface {
	// Create stores a new metric record in the repository.
	Create(metric entities.Metrics) error

	// Get retrieves a metric record based on the provided key and type.
	// It returns the metric and an error if the operation fails.
	Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error)

	// GetAll retrieves all metric records from the repository.
	// It returns a pointer to MetricsStorage and an error if the operation fails.
	GetAll() (*storage.MetricsStorage, error)

	// GetAllByType retrieves all metric records of a specific type from the repository.
	// It returns a map of metric names to metrics and an error if the operation fails.
	GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)

	// StoreMetricsBatch stores a batch of metric records in the repository.
	// It returns an error if the operation fails.
	StoreMetricsBatch(metrics []entities.Metrics) error
}

// Repository is a struct that holds the MetricsRepository interface.
// It provides a unified way to access metric data storage functionalities.
type Repository struct {
	MetricsRepository MetricsRepository // The underlying metrics repository implementation.
}

// NewRepository creates a new instance of Repository.
// It takes a storage.Storage instance to initialize the MetricsRepository.
func NewRepository(store storage.Storage) *Repository {
	return &Repository{
		MetricsRepository: NewMetricsMemRepository(store),
	}
}
