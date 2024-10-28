// Package repositories provides implementations for the metrics collection
// repositories, including in-memory storage for metrics.
//
// This file contains the MetricsMemRepository, which facilitates operations
// for creating, retrieving, and storing metrics in the underlying storage.
package repositories

import (
	"errors"
	"fmt"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// MetricsMemRepository is an in-memory repository for metrics that interacts
// with a storage interface. It provides methods for creating, retrieving,
// and storing metrics data.
type MetricsMemRepository struct {
	store storage.Storage // The underlying storage interface for metrics.
}

// NewMetricsMemRepository creates a new instance of MetricsMemRepository
// with the provided storage interface.
func NewMetricsMemRepository(store storage.Storage) *MetricsMemRepository {
	return &MetricsMemRepository{store: store}
}

// Create saves a new metric record in the repository. It returns an error
// if the creation fails.
func (cmr *MetricsMemRepository) Create(metric entities.Metrics) error {
	err := cmr.store.CreateRecord(metric)
	if err != nil {
		return fmt.Errorf("failed to create the record: %w", err)
	}

	return nil
}

// Get retrieves a metric record by its key and type. It returns an error
// if the item is not found or if retrieval fails.
func (cmr *MetricsMemRepository) Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	record, err := cmr.store.GetRecord(key, mType)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Metrics{}, fmt.Errorf("item with key %s was not found: %w", key, err)
		}

		return entities.Metrics{}, fmt.Errorf("failed to get the item: %w", err)
	}

	return record, nil
}

// GetAll retrieves all metrics records from the repository. It returns a
// MetricsStorage object or an error if retrieval fails.
func (cmr *MetricsMemRepository) GetAll() (*storage.MetricsStorage, error) {
	store, err := cmr.store.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics: %w", err)
	}

	return store, nil
}

// GetAllByType retrieves all metrics of a specific type from the repository.
// It returns a map of metric names to their corresponding metrics or an error
// if retrieval fails.
func (cmr *MetricsMemRepository) GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	metrics, err := cmr.store.GetAllRecordsByType(mType)
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics: %w", err)
	}

	return metrics, nil
}

// StoreMetricsBatch stores a batch of metric records in the repository.
// It returns an error if the batch storage operation fails.
func (cmr *MetricsMemRepository) StoreMetricsBatch(metrics []entities.Metrics) error {
	if err := cmr.store.StoreMetricsBatch(metrics); err != nil {
		return fmt.Errorf("mem storage error: %w", err)
	}

	return nil
}
