package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type MetricsCollectionMemRepository struct {
	store  *storage.MetricsCollection
	logger *slog.Logger
}

func NewMetricsCollectionMemRepository(collection *storage.MetricsCollection,
	logger *slog.Logger) *MetricsCollectionMemRepository {
	return &MetricsCollectionMemRepository{store: collection, logger: logger}
}

func (m *MetricsCollectionMemRepository) Store(stats *runtime.MemStats) error {
	//nolint:exhaustive // entities.PollCount is of type Counter
	gaugeMetrics := map[entities.MetricName]entities.Gauge{
		entities.RandomValue:   entities.Gauge(rand.Float64()),
		entities.Alloc:         entities.Gauge(stats.Alloc),
		entities.BuckHashSys:   entities.Gauge(stats.BuckHashSys),
		entities.Frees:         entities.Gauge(stats.Frees),
		entities.GCCPUFraction: entities.Gauge(stats.GCCPUFraction),
		entities.GCSys:         entities.Gauge(stats.GCSys),
		entities.HeapAlloc:     entities.Gauge(stats.HeapAlloc),
		entities.HeapIdle:      entities.Gauge(stats.HeapIdle),
		entities.HeapInuse:     entities.Gauge(stats.HeapInuse),
		entities.HeapObjects:   entities.Gauge(stats.HeapObjects),
		entities.HeapReleased:  entities.Gauge(stats.HeapReleased),
		entities.HeapSys:       entities.Gauge(stats.HeapSys),
		entities.LastGC:        entities.Gauge(stats.LastGC),
		entities.Lookups:       entities.Gauge(stats.Lookups),
		entities.MCacheInuse:   entities.Gauge(stats.MCacheInuse),
		entities.MCacheSys:     entities.Gauge(stats.MCacheSys),
		entities.MSpanInuse:    entities.Gauge(stats.MSpanInuse),
		entities.MSpanSys:      entities.Gauge(stats.MSpanSys),
		entities.Mallocs:       entities.Gauge(stats.Mallocs),
		entities.NextGC:        entities.Gauge(stats.NextGC),
		entities.NumForcedGC:   entities.Gauge(stats.NumForcedGC),
		entities.NumGC:         entities.Gauge(stats.NumGC),
		entities.OtherSys:      entities.Gauge(stats.OtherSys),
		entities.PauseTotalNs:  entities.Gauge(stats.PauseTotalNs),
		entities.StackInuse:    entities.Gauge(stats.StackInuse),
		entities.StackSys:      entities.Gauge(stats.StackSys),
		entities.Sys:           entities.Gauge(stats.Sys),
		entities.TotalAlloc:    entities.Gauge(stats.TotalAlloc),
	}

	if err := m.store.StoreGauge(gaugeMetrics); err != nil {
		return fmt.Errorf("failed to store the metrics: %w", err)
	}

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
