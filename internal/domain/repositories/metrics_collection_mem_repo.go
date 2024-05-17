package repositories

import (
	"math/rand"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type MetricsCollectionMemRepository struct {
	store *storage.MetricsCollection
}

func NewMetricsCollectionMemRepository(collection *storage.MetricsCollection) *MetricsCollectionMemRepository {
	return &MetricsCollectionMemRepository{store: collection}
}

func (m *MetricsCollectionMemRepository) Store(stats *runtime.MemStats) {
	m.store.Mu.Lock()
	defer m.store.Mu.Unlock()

	if len(m.store.Collection.CounterMetrics) == 0 {
		m.store.Collection.CounterMetrics = append(m.store.Collection.CounterMetrics,
			entities.CounterMetric{Name: entities.PollCount, Value: 0},
		)
	}

	for i, v := range m.store.Collection.CounterMetrics {
		if v.Name == entities.PollCount {
			m.store.Collection.CounterMetrics[i].Value++
		}
	}

	// Gauge Metrics
	gaugeMetrics := []entities.GaugeMetric{
		{Name: entities.RandomValue, Value: entities.Gauge(rand.Float64())},
		{Name: entities.Alloc, Value: entities.Gauge(stats.Alloc)},
		{Name: entities.BuckHashSys, Value: entities.Gauge(stats.BuckHashSys)},
		{Name: entities.Frees, Value: entities.Gauge(stats.Frees)},
		{Name: entities.GCCPUFraction, Value: entities.Gauge(stats.GCCPUFraction)},
		{Name: entities.GCSys, Value: entities.Gauge(stats.GCSys)},
		{Name: entities.HeapAlloc, Value: entities.Gauge(stats.HeapAlloc)},
		{Name: entities.HeapIdle, Value: entities.Gauge(stats.HeapIdle)},
		{Name: entities.HeapInuse, Value: entities.Gauge(stats.HeapInuse)},
		{Name: entities.HeapObjects, Value: entities.Gauge(stats.HeapObjects)},
		{Name: entities.HeapReleased, Value: entities.Gauge(stats.HeapReleased)},
		{Name: entities.HeapSys, Value: entities.Gauge(stats.HeapSys)},
		{Name: entities.LastGC, Value: entities.Gauge(stats.LastGC)},
		{Name: entities.Lookups, Value: entities.Gauge(stats.Lookups)},
		{Name: entities.MCacheInuse, Value: entities.Gauge(stats.MCacheInuse)},
		{Name: entities.MCacheSys, Value: entities.Gauge(stats.MCacheSys)},
		{Name: entities.MSpanInuse, Value: entities.Gauge(stats.MSpanInuse)},
		{Name: entities.MSpanSys, Value: entities.Gauge(stats.MSpanSys)},
		{Name: entities.Mallocs, Value: entities.Gauge(stats.Mallocs)},
		{Name: entities.NextGC, Value: entities.Gauge(stats.NextGC)},
		{Name: entities.NumForcedGC, Value: entities.Gauge(stats.NumForcedGC)},
		{Name: entities.NumGC, Value: entities.Gauge(stats.NumGC)},
		{Name: entities.OtherSys, Value: entities.Gauge(stats.OtherSys)},
		{Name: entities.PauseTotalNs, Value: entities.Gauge(stats.PauseTotalNs)},
		{Name: entities.StackInuse, Value: entities.Gauge(stats.StackInuse)},
		{Name: entities.StackSys, Value: entities.Gauge(stats.StackSys)},
		{Name: entities.Sys, Value: entities.Gauge(stats.Sys)},
		{Name: entities.TotalAlloc, Value: entities.Gauge(stats.TotalAlloc)},
	}

	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, gaugeMetrics...)
}

func (m *MetricsCollectionMemRepository) GetAll() *entities.MetricsCollection {
	m.store.Mu.Lock()
	defer m.store.Mu.Unlock()
	collection := m.store.Collection
	newCounterMetrics := make([]entities.CounterMetric, len(collection.CounterMetrics))
	newGaugeMetrics := make([]entities.GaugeMetric, len(collection.GaugeMetrics))

	_ = copy(newCounterMetrics, collection.CounterMetrics)
	_ = copy(newGaugeMetrics, collection.GaugeMetrics)

	newMetricsCollection := entities.MetricsCollection{
		CounterMetrics: newCounterMetrics,
		GaugeMetrics:   newGaugeMetrics,
	}

	return &newMetricsCollection
}

func (m *MetricsCollectionMemRepository) Clear() {
	m.store.Mu.Lock()
	defer m.store.Mu.Unlock()

	m.store.Collection.GaugeMetrics = m.store.Collection.GaugeMetrics[:0]
	m.store.Collection.CounterMetrics = m.store.Collection.CounterMetrics[:0]
}
