package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

var ErrNotFound = errors.New("item not found")

type MetricsStorage struct {
	Counter map[entities.MetricName]entities.Counter `json:"counter"`
	Gauge   map[entities.MetricName]entities.Gauge   `json:"gauge"`
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}
}

type MemStorage struct {
	logger *slog.Logger
	MetricsStorage
	mu sync.Mutex
}

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

func (ms *MemStorage) CreateRecord(metrics entities.Metrics) error {
	ms.logger.DebugContext(context.Background(), fmt.Sprintf("creating %s record", metrics.MType))

	switch entities.MetricType(metrics.MType) {
	case entities.CounterMetricName:
		if err := ms.createCounterRecord(metrics); err != nil {
			return fmt.Errorf("store couter: %w", err)
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

func (ms *MemStorage) createCounterRecord(metric entities.Metrics) error {
	if ms.Counter == nil {
		return errors.New("mem not initialized")
	}

	ms.mu.Lock()
	_, ok := ms.Counter[entities.MetricName(metric.ID)]
	if !ok {
		ms.Counter[entities.MetricName(metric.ID)] = entities.Counter(*metric.Delta)
	} else {
		ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta)
	}
	ms.mu.Unlock()

	return nil
}

func (ms *MemStorage) createGaugeRecord(metric entities.Metrics) error {
	if ms.Gauge == nil {
		return errors.New("gauge memory not initialized")
	}

	ms.mu.Lock()
	ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value)
	ms.mu.Unlock()

	return nil
}

func (ms *MemStorage) GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.logger.DebugContext(context.Background(), fmt.Sprintf("retrieving %s[%s] record", mType, mName))

	switch mType {
	case entities.CounterMetricName:
		m, ok := ms.Counter[mName]
		if !ok {
			return entities.Metrics{}, ErrNotFound
		}
		val := int64(m)
		return entities.Metrics{
			ID:    string(mName),
			MType: string(mType),
			Delta: &val,
		}, nil
	case entities.GaugeMetricName:
		m, ok := ms.Gauge[mName]
		if !ok {
			return entities.Metrics{}, ErrNotFound
		}
		val := float64(m)
		return entities.Metrics{
			ID:    string(mName),
			MType: string(mType),
			Value: &val,
		}, nil
	}

	return entities.Metrics{}, ErrNotFound
}

func (ms *MemStorage) GetAllRecords() (*MetricsStorage, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Counter == nil {
		return nil, errors.New("storage not initialized")
	}

	copyCounterMap := make(map[entities.MetricName]entities.Counter)
	copyGaugeMap := make(map[entities.MetricName]entities.Gauge)
	for k, v := range ms.Counter {
		copyCounterMap[k] = v
	}

	for k, v := range ms.Gauge {
		copyGaugeMap[k] = v
	}

	return &MetricsStorage{
		Counter: copyCounterMap,
		Gauge:   copyGaugeMap,
	}, nil
}

func (ms *MemStorage) GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics,
	error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	copyCounterMap := make(map[entities.MetricName]entities.Metrics)

	switch mType {
	case entities.CounterMetricName:
		for k, v := range ms.Counter {
			val := int64(v)
			copyCounterMap[k] = entities.Metrics{ID: string(k), MType: string(mType), Delta: &val}
		}
	case entities.GaugeMetricName:
		for k, v := range ms.Counter {
			val := float64(v)
			copyCounterMap[k] = entities.Metrics{ID: string(k), MType: string(mType), Value: &val}
		}
	}

	return copyCounterMap, nil
}

func (ms *MemStorage) StoreMetricsBatch(metrics []entities.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, metric := range metrics {
		switch entities.MetricType(metric.MType) {
		case entities.GaugeMetricName:
			ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value)
		case entities.CounterMetricName:
			ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta)
		}
	}

	return nil
}

func (ms *MemStorage) Close(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.MetricsStorage = MetricsStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}

	return nil
}
