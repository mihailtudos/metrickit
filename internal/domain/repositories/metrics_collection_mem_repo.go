package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"math/rand"
	"runtime"
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
		m.store.Collection.CounterMetrics = append(m.store.Collection.CounterMetrics, entities.CounterMetric{Name: entities.PollCount, Value: 0})
	}

	for i, v := range m.store.Collection.CounterMetrics {
		if v.Name == entities.PollCount {
			m.store.Collection.CounterMetrics[i].Value += 1
		}
	}

	// Gauge Metrics
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.RandomValue, Value: entities.Gauge(rand.Float64())})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.Alloc, Value: entities.Gauge(stats.Alloc)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.BuckHashSys, Value: entities.Gauge(stats.BuckHashSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.Frees, Value: entities.Gauge(stats.Frees)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.GCCPUFraction, Value: entities.Gauge(stats.GCCPUFraction)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.GCSys, Value: entities.Gauge(stats.GCSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapAlloc, Value: entities.Gauge(stats.HeapAlloc)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapIdle, Value: entities.Gauge(stats.HeapIdle)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapInuse, Value: entities.Gauge(stats.HeapInuse)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapObjects, Value: entities.Gauge(stats.HeapObjects)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapReleased, Value: entities.Gauge(stats.HeapReleased)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapSys, Value: entities.Gauge(stats.HeapSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.LastGC, Value: entities.Gauge(stats.LastGC)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.Lookups, Value: entities.Gauge(stats.Lookups)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.MCacheInuse, Value: entities.Gauge(stats.MCacheInuse)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.MCacheSys, Value: entities.Gauge(stats.MCacheSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.MSpanInuse, Value: entities.Gauge(stats.MSpanInuse)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.MSpanSys, Value: entities.Gauge(stats.MSpanSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.Mallocs, Value: entities.Gauge(stats.Mallocs)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.NextGC, Value: entities.Gauge(stats.NextGC)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.NumForcedGC, Value: entities.Gauge(stats.NumForcedGC)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.NumGC, Value: entities.Gauge(stats.NumGC)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.OtherSys, Value: entities.Gauge(stats.OtherSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.PauseTotalNs, Value: entities.Gauge(stats.PauseTotalNs)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.StackInuse, Value: entities.Gauge(stats.StackInuse)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.StackSys, Value: entities.Gauge(stats.StackSys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.Sys, Value: entities.Gauge(stats.Sys)})
	m.store.Collection.GaugeMetrics = append(m.store.Collection.GaugeMetrics, entities.GaugeMetric{Name: entities.TotalAlloc, Value: entities.Gauge(stats.TotalAlloc)})
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
