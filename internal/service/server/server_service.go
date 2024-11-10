// Package server provides the service layer for handling metrics operations.
// It defines the Metrics interface and implements it through the Service struct.
package server

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// Metrics defines the interface for metrics operations.
// It includes methods for creating, retrieving, and storing metrics.
//
//go:generate mockgen -destination=mocks/mock_metrics.go -package=mocks github.com/mihailtudos/metrickit/internal/service/server Metrics
type Metrics interface {
	// Create adds a new metric to the storage.
	Create(metric entities.Metrics) error

	// Get retrieves a metric by its name and type.
	Get(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error)

	// GetAll retrieves all metrics from the storage.
	GetAll() (*storage.MetricsStorage, error)

	// GetAllByType retrieves all metrics of a specific type from the storage.
	GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)

	// StoreMetricsBatch stores a batch of metrics in the storage.
	StoreMetricsBatch(metrics []entities.Metrics) error
}

// Service provides methods for managing metrics.
// It implements the Metrics interface and uses a repository to perform operations.
type Service struct {
	MetricsService Metrics
}

// NewMetricsService creates a new instance of the Service struct.
// It initializes the metrics service with the provided repository and logger.
func NewMetricsService(repository *repositories.Repository, logger *slog.Logger) *MetricsService {
	return NewMetricService(repository.MetricsRepository, logger) // Initialize the metrics service.
}
