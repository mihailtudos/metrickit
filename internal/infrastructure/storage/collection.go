package storage

import (
	"errors"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type MetricsCollection struct {
	Collection *entities.MetricsCollection
	mu         sync.Mutex
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		mu:         sync.Mutex{},
		Collection: entities.NewMetricsCollection(),
	}
}

func (m *MetricsCollection) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Collection = entities.NewMetricsCollection()
}

func (m *MetricsCollection) StoreCounter() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Collection.CounterMetrics == nil {
		return errors.New("counter store not initialized")
	}

	if _, ok := m.Collection.CounterMetrics[entities.PollCount]; !ok {
		m.Collection.GaugeMetrics[entities.PollCount] = 0
	}

	m.Collection.CounterMetrics[entities.PollCount]++

	return nil
}

func (m *MetricsCollection) StoreGauge(gauges map[entities.MetricName]entities.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Collection.GaugeMetrics == nil {
		return errors.New("gauge store not initialized")
	}

	for k, v := range gauges {
		m.Collection.GaugeMetrics[k] = v
	}

	return nil
}

func (m *MetricsCollection) GetCounterCollection() (map[entities.MetricName]entities.Counter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	collection := m.Collection.CounterMetrics

	newCounterMetrics := make(map[entities.MetricName]entities.Counter, len(collection))
	copyMap(collection, newCounterMetrics)
	return newCounterMetrics, nil
}

func (m *MetricsCollection) GetGaugeCollection() (map[entities.MetricName]entities.Gauge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	collection := m.Collection.GaugeMetrics

	newCounterMetrics := make(map[entities.MetricName]entities.Gauge, len(collection))
	copyMap(collection, newCounterMetrics)
	return newCounterMetrics, nil
}

func copyMap[T any](src, dst map[entities.MetricName]T) {
	for key, value := range src {
		dst[key] = value
	}
}
