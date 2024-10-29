// Package storage provides mechanisms for storing and managing metrics.
// It defines memory-based storage for metrics collections and allows for
// concurrent access to store and retrieve metrics data.
package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

// ErrNotFound is returned when a requested item cannot be found in storage.
var ErrNotFound = errors.New("item not found")

// MetricsStorage holds collections of counter and gauge metrics.
type MetricsStorage struct {
	Counter map[entities.MetricName]entities.Counter `json:"counter"` // Collection of counter metrics.
	Gauge   map[entities.MetricName]entities.Gauge   `json:"gauge"`   // Collection of gauge metrics.
}

// NewMetricsStorage initializes a new MetricsStorage instance.
func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}
}

// MemStorage provides an in-memory implementation of metrics storage,
// supporting concurrent access and operations.
type MemStorage struct {
	logger         *slog.Logger // Logger for logging events and errors.
	MetricsStorage              // Embedding MetricsStorage for storage functionalities.
	mu             sync.Mutex   // Mutex for synchronizing access to storage.
}

// NewMemStorage creates a new MemStorage instance with logging capabilities.
func NewMemStorage(logger *slog.Logger) (*MemStorage, error) {
	logger.DebugContext(context.Background(), "created mem storage")

	ms := &MemStorage{
		mu: sync.Mutex{},
		MetricsStorage: MetricsStorage{
			Counter: make(map[entities.MetricName]entities.Counter),
			Gauge:   make(map[entities.MetricName]entities.Gauge),
		},
		logger: logger,
	}

	return ms, nil
}

// CreateRecord stores a new metrics record in memory,
// determining whether it is a counter or gauge metric.
func (ms *MemStorage) CreateRecord(metrics entities.Metrics) error {
	ms.logger.DebugContext(context.Background(), fmt.Sprintf("creating %s record", metrics.MType))

	switch entities.MetricType(metrics.MType) {
	case entities.CounterMetricName:
		if err := ms.createCounterRecord(metrics); err != nil {
			return fmt.Errorf("store counter: %w", err)
		}
		return nil
	case entities.GaugeMetricName:
		if err := ms.createGaugeRecord(metrics); err != nil {
			return fmt.Errorf("store gauge: %w", err)
		}
		return nil
	default:
		return errors.New("store: unsupported record type " + metrics.MType)
	}
}

// createCounterRecord adds or updates a counter metric in memory.
func (ms *MemStorage) createCounterRecord(metric entities.Metrics) error {
	if ms.Counter == nil {
		return errors.New("mem not initialized") // Error if counter store is uninitialized.
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	_, ok := ms.Counter[entities.MetricName(metric.ID)]
	if !ok {
		ms.Counter[entities.MetricName(metric.ID)] = entities.Counter(*metric.Delta) // Initialize counter if not present.
	} else {
		ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta) // Increment existing counter.
	}

	return nil
}

// createGaugeRecord adds a gauge metric in memory.
func (ms *MemStorage) createGaugeRecord(metric entities.Metrics) error {
	if ms.Gauge == nil {
		return errors.New("gauge memory not initialized") // Error if gauge store is uninitialized.
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value) // Store gauge metric.

	return nil
}

// GetRecord retrieves a specific metrics record by name and type.
func (ms *MemStorage) GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.logger.DebugContext(context.Background(), fmt.Sprintf("retrieving %s[%s] record", mType, mName))

	switch mType {
	case entities.CounterMetricName:
		m, ok := ms.Counter[mName]
		if !ok {
			return entities.Metrics{}, ErrNotFound // Return error if metric not found.
		}
		val := int64(m)
		return entities.Metrics{
			ID:    string(mName),
			MType: string(mType),
			Delta: &val, // Return counter value.
		}, nil
	case entities.GaugeMetricName:
		m, ok := ms.Gauge[mName]
		if !ok {
			return entities.Metrics{}, ErrNotFound // Return error if metric not found.
		}
		val := float64(m)
		return entities.Metrics{
			ID:    string(mName),
			MType: string(mType),
			Value: &val, // Return gauge value.
		}, nil
	}

	return entities.Metrics{}, ErrNotFound // Unsupported metric type.
}

// GetAllRecords retrieves all metrics records in storage.
func (ms *MemStorage) GetAllRecords() (*MetricsStorage, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Counter == nil {
		return nil, errors.New("storage not initialized") // Error if storage is uninitialized.
	}

	copyCounterMap := make(map[entities.MetricName]entities.Counter)
	copyGaugeMap := make(map[entities.MetricName]entities.Gauge)
	for k, v := range ms.Counter {
		copyCounterMap[k] = v // Copy counter metrics for safe return.
	}

	for k, v := range ms.Gauge {
		copyGaugeMap[k] = v // Copy gauge metrics for safe return.
	}

	return &MetricsStorage{
		Counter: copyCounterMap,
		Gauge:   copyGaugeMap,
	}, nil
}

// GetAllRecordsByType retrieves all metrics records of a specified type.
func (ms *MemStorage) GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	copyMetricsMap := make(map[entities.MetricName]entities.Metrics)

	switch mType {
	case entities.CounterMetricName:
		for k, v := range ms.Counter {
			val := int64(v)
			copyMetricsMap[k] = entities.Metrics{ID: string(k), MType: string(mType), Delta: &val} // Store counter metrics.
		}
	case entities.GaugeMetricName:
		for k, v := range ms.Gauge {
			val := float64(v)
			copyMetricsMap[k] = entities.Metrics{ID: string(k), MType: string(mType), Value: &val} // Store gauge metrics.
		}
	}

	return copyMetricsMap, nil
}

// StoreMetricsBatch stores a batch of metrics records in memory.
func (ms *MemStorage) StoreMetricsBatch(metrics []entities.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, metric := range metrics {
		switch entities.MetricType(metric.MType) {
		case entities.GaugeMetricName:
			ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value) // Store gauge metric.
		case entities.CounterMetricName:
			ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta) // Increment counter metric.
		}
	}

	return nil
}

// Close resets the metrics storage, clearing all metrics data.
func (ms *MemStorage) Close(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.MetricsStorage = MetricsStorage{ // Reinitialize metrics storage.
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}

	return nil
}
