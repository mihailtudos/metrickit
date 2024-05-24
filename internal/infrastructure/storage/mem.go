package storage

import (
	"errors"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

var ErrNotFound = errors.New("item not found")

type MemStorage struct {
	Counter map[entities.MetricName]entities.Counter
	Gauge   map[entities.MetricName]entities.Gauge
	mu      sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mu:      sync.Mutex{},
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}
}

func (ms *MemStorage) CreateCounterRecord(metric entities.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	_, ok := ms.Counter[entities.MetricName(metric.ID)]
	if !ok {
		ms.Counter[entities.MetricName(metric.ID)] = entities.Counter(*metric.Delta)
	} else {
		ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta)
	}

	return nil
}

func (ms *MemStorage) CreateGaugeRecord(metric entities.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value)
	return nil
}

func (ms *MemStorage) GetGaugeRecord(key entities.MetricName) (entities.Gauge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Gauge == nil {
		return entities.Gauge(0), errors.New("gauge storage not initiated")
	}

	v, ok := ms.Gauge[key]
	if !ok {
		return entities.Gauge(0), ErrNotFound
	}

	return v, nil
}

func (ms *MemStorage) GetCounterRecord(key entities.MetricName) (entities.Counter, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Counter == nil {
		return entities.Counter(1), errors.New("counter storage not initiated")
	}

	v, ok := ms.Counter[key]
	if !ok {
		return entities.Counter(0), ErrNotFound
	}

	return v, nil
}

func (ms *MemStorage) GetAllGaugeRecords() (map[entities.MetricName]entities.Gauge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Gauge == nil {
		return nil, errors.New("storage not initialized")
	}

	copyMap := make(map[entities.MetricName]entities.Gauge)
	for k, v := range ms.Gauge {
		copyMap[k] = v
	}

	return copyMap, nil
}

func (ms *MemStorage) GetAllCounterRecords() (map[entities.MetricName]entities.Counter, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Counter == nil {
		return nil, errors.New("storage not initialized")
	}

	copyMap := make(map[entities.MetricName]entities.Counter)
	for k, v := range ms.Counter {
		copyMap[k] = v
	}

	return copyMap, nil
}
