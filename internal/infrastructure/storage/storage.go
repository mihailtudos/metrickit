// Package storage provides an abstraction for various storage implementations
// for metrics, including in-memory, file-based, and PostgreSQL storage.
package storage

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

// Storage defines the methods for storing and retrieving metrics records.
// Any type that implements this interface can be used for metrics storage.
type Storage interface {
	// CreateRecord adds a new metrics record to the storage.
	CreateRecord(metrics entities.Metrics) error

	// GetRecord retrieves a specific metrics record by name and type.
	GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error)

	// GetAllRecords returns all metrics records stored.
	GetAllRecords() (*MetricsStorage, error)

	// GetAllRecordsByType retrieves all metrics records of a specified type.
	GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)

	// StoreMetricsBatch stores a batch of metrics records in the storage.
	StoreMetricsBatch(metrics []entities.Metrics) error

	// Close gracefully shuts down the storage, releasing any resources.
	Close(ctx context.Context) error
}

// NewStorage creates a new storage instance based on the provided parameters.
// It prioritizes PostgreSQL storage if a database connection is provided,
// then file storage if a valid storeInterval is specified, and finally,
// defaults to in-memory storage.
func NewStorage(db *pgxpool.Pool, logger *slog.Logger, storeInterval int, storePath string) (Storage, error) {
	if db != nil {
		return NewPostgresStorage(db, logger) // Create PostgreSQL storage if db is not nil.
	}

	if storeInterval >= 0 {
		return NewFileStorage(logger, storeInterval, storePath) // Create file storage if storeInterval is valid.
	}

	return NewMemStorage(logger) // Default to in-memory storage.
}
