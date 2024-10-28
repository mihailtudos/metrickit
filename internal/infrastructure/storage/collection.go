// Package storage provides the storage mechanism for metrics collections.
// It defines the MetricsCollection struct for managing and storing metrics data.
package storage

import (
	"errors"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

// MetricsCollection manages a collection of metrics, providing concurrency-safe operations
// for storing and retrieving counter and gauge metrics.
type MetricsCollection struct {
	collection *entities.MetricsCollection // The underlying metrics collection.
	mu         sync.Mutex                  // Mutex to ensure concurrent access safety.
}

// NewMetricsCollection creates and initializes a new MetricsCollection instance.
func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		mu:         sync.Mutex{},
		collection: entities.NewMetricsCollection(),
	}
}

// Clear resets the metrics collection, removing all stored metrics.
func (m *MetricsCollection) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collection = entities.NewMetricsCollection() // Reinitialize the metrics collection.
}

// StoreCounter increments the poll count in the metrics collection.
// It initializes the counter if it has not been done yet.
func (m *MetricsCollection) StoreCounter() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.collection.CounterMetrics == nil {
		return errors.New("counter store not initialized") // Error if counter store is uninitialized.
	}

	if _, ok := m.collection.CounterMetrics[entities.PollCount]; !ok {
		m.collection.CounterMetrics[entities.PollCount] = 0 // Initialize the counter if not present.
	}

	m.collection.CounterMetrics[entities.PollCount]++ // Increment the counter.

	return nil
}

// StoreGauge stores multiple gauge metrics in the metrics collection.
func (m *MetricsCollection) StoreGauge(gauges map[entities.MetricName]entities.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.collection.GaugeMetrics == nil {
		return errors.New("gauge store not initialized") // Error if gauge store is uninitialized.
	}

	for k, v := range gauges {
		m.collection.GaugeMetrics[k] = v // Store each gauge metric.
	}

	return nil
}

// GetCounterCollection retrieves a copy of the counter metrics collection.
func (m *MetricsCollection) GetCounterCollection() (map[entities.MetricName]entities.Counter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.collection.CounterMetrics
	newCounterMetrics := make(map[entities.MetricName]entities.Counter, len(collection))
	copyMap(collection, newCounterMetrics) // Copy the counter metrics for safe return.
	return newCounterMetrics, nil
}

// GetCounterMetric retrieves a specific counter metric by name.
func (m *MetricsCollection) GetCounterMetric(mName entities.MetricName) (entities.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.collection.CounterMetrics[mName]
	if !ok {
		return entities.Metrics{}, ErrNotFound // Return error if metric not found.
	}

	val := int64(v)
	return entities.Metrics{ID: string(mName), MType: string(entities.CounterMetricName), Delta: &val}, nil // Return the metric.
}

// GetGaugeCollection retrieves a copy of the gauge metrics collection.
func (m *MetricsCollection) GetGaugeCollection() (map[entities.MetricName]entities.Gauge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.collection.GaugeMetrics
	newGaugeMetrics := make(map[entities.MetricName]entities.Gauge, len(collection))
	copyMap(collection, newGaugeMetrics) // Copy the gauge metrics for safe return.
	return newGaugeMetrics, nil
}

// copyMap copies elements from the source map to the destination map.
func copyMap[T any](src, dst map[entities.MetricName]T) {
	for key, value := range src {
		dst[key] = value
	}
}
