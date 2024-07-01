package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"log/slog"
)

type MetricsCollectionMemRepository struct {
	store  *storage.MetricsCollection
	logger *slog.Logger
}

func NewMetricsCollectionMemRepository(collection *storage.MetricsCollection,
	logger *slog.Logger) *MetricsCollectionMemRepository {
	return &MetricsCollectionMemRepository{store: collection, logger: logger}
}

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
