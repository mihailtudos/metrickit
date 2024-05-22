package storage

import (
	"errors"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

var ErrNotFound = errors.New("item not found")

type MemStorage struct {
	Counter map[string]entities.Counter
	Gauge   map[string]entities.Gauge
	mu      sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mu:      sync.Mutex{},
		Counter: make(map[string]entities.Counter),
		Gauge:   make(map[string]entities.Gauge),
	}
}

func (ms *MemStorage) CreateCounterRecord(key string, record entities.Counter) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	_, ok := ms.Counter[key]
	if !ok {
		ms.Counter[key] = record
	} else {
		ms.Counter[key] += record
	}

	return nil
}

func (ms *MemStorage) CreateGaugeRecord(key string, record entities.Gauge) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Gauge[key] = record
	return nil
}

func (ms *MemStorage) GetGaugeRecord(key string) (entities.Gauge, error) {
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

func (ms *MemStorage) GetCounterRecord(key string) (entities.Counter, error) {
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

func (ms *MemStorage) GetAllGaugeRecords() (map[string]entities.Gauge, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Gauge == nil {
		return nil, errors.New("storage not initialized")
	}

	copyMap := make(map[string]entities.Gauge)
	for k, v := range ms.Gauge {
		copyMap[k] = v
	}

	return copyMap, nil
}

func (ms *MemStorage) GetAllCounterRecords() (map[string]entities.Counter, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Counter == nil {
		return nil, errors.New("storage not initialized")
	}

	copyMap := make(map[string]entities.Counter)
	for k, v := range ms.Counter {
		copyMap[k] = v
	}

	return copyMap, nil
}
