package storage

import (
	"errors"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type MetricsCollection struct {
	collection *entities.MetricsCollection
	mu         sync.Mutex
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		mu:         sync.Mutex{},
		collection: entities.NewMetricsCollection(),
	}
}

func (m *MetricsCollection) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collection = entities.NewMetricsCollection()
}

func (m *MetricsCollection) StoreCounter() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.collection.CounterMetrics == nil {
		return errors.New("counter store not initialized")
	}

	if _, ok := m.collection.CounterMetrics[entities.PollCount]; !ok {
		m.collection.CounterMetrics[entities.PollCount] = 0
	}

	m.collection.CounterMetrics[entities.PollCount]++

	return nil
}

func (m *MetricsCollection) StoreGauge(gauges map[entities.MetricName]entities.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.collection.GaugeMetrics == nil {
		return errors.New("gauge store not initialized")
	}

	for k, v := range gauges {
		m.collection.GaugeMetrics[k] = v
	}

	return nil
}

func (m *MetricsCollection) GetCounterCollection() (map[entities.MetricName]entities.Counter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	collection := m.collection.CounterMetrics

	newCounterMetrics := make(map[entities.MetricName]entities.Counter, len(collection))
	copyMap(collection, newCounterMetrics)
	return newCounterMetrics, nil
}

func (m *MetricsCollection) GetCounterMetric(mName entities.MetricName) (entities.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.collection.CounterMetrics[mName]
	if !ok {
		return entities.Metrics{}, ErrNotFound
	}

	val := int64(v)
	return entities.Metrics{ID: string(mName), MType: string(entities.CounterMetricName), Delta: &val}, nil
}

func (m *MetricsCollection) GetGaugeCollection() (map[entities.MetricName]entities.Gauge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	collection := m.collection.GaugeMetrics

	newCounterMetrics := make(map[entities.MetricName]entities.Gauge, len(collection))
	copyMap(collection, newCounterMetrics)
	return newCounterMetrics, nil
}

func copyMap[T any](src, dst map[entities.MetricName]T) {
	for key, value := range src {
		dst[key] = value
	}
}
