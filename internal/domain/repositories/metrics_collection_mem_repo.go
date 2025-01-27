// Package repositories provides implementations for the metrics collection
// repositories, including in-memory storage for metrics.
//
// This file contains the MetricsCollectionMemRepository, which stores
// metrics in a given storage collection and allows for the retrieval
// of both gauge and counter metrics.
package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

// MetricsCollectionMemRepository is an in-memory repository for storing
// and retrieving metrics. It uses a MetricsCollection as its underlying
// storage and provides methods for storing gauge metrics and retrieving
// both gauge and counter metrics.
type MetricsCollectionMemRepository struct {
	store  *storage.MetricsCollection // The underlying storage for metrics.
	logger *slog.Logger               // Logger for logging operations.
}

// NewMetricsCollectionMemRepository creates a new instance of
// MetricsCollectionMemRepository with the provided storage collection
// and logger.
func NewMetricsCollectionMemRepository(collection *storage.MetricsCollection,
	logger *slog.Logger) *MetricsCollectionMemRepository {
	return &MetricsCollectionMemRepository{store: collection, logger: logger}
}

// Store saves the given gauge metrics in the repository and updates
// the poll counter metric. It logs the updated metrics after storage.
// Returns an error if storing metrics fails.
func (m *MetricsCollectionMemRepository) Store(gaugeMetrics map[entities.MetricName]entities.Gauge) error {
	if err := m.store.StoreGauge(gaugeMetrics); err != nil {
		return fmt.Errorf("failed to store the metrics: %w", err)
	}

	// updating poll counter
	if err := m.store.StoreCounter(); err != nil {
		return fmt.Errorf("failed to store the metrics: %w", err)
	}

	pc, err := m.store.GetCounterMetric(entities.PollCount)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("store metric: %w", err)
		}
		return fmt.Errorf("failed to get counter metric: %w", err)
	}

	m.logger.DebugContext(
		context.Background(),
		"updated metrics",
		slog.Int64("PoolCount", *pc.Delta))
	return nil
}

// GetAll retrieves all metrics from the repository, including both
// gauge and counter metrics. Returns a MetricsCollection containing
// the retrieved metrics or an error if retrieval fails.
func (m *MetricsCollectionMemRepository) GetAll() (*entities.MetricsCollection, error) {
	counters, err := m.store.GetCounterCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the Counter collection: %w", err)
	}

	gauges, err := m.store.GetGaugeCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the Gauge collection: %w", err)
	}

	newMetricsCollection := entities.MetricsCollection{
		CounterMetrics: counters,
		GaugeMetrics:   gauges,
	}

	return &newMetricsCollection, nil
}
