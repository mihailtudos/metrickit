package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mihailtudos/metrickit/internal/config"
	"os"
	"sync"
	"time"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

var ErrNotFound = errors.New("item not found")
var ownerFilePerm os.FileMode = 0o600

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
	MetricsStorage
	stopSaveChan chan struct{}
	cfg          *config.ServerConfig
	// TODO(SSH): seems, that adding writing metrics to file should not really change the in-memory store implementation
	// it would be better to:
	// 1. Define a `FileStorage` type, which is different from `MemStorage`, but wraps its functionality (you can embed the `MemStorage` into it)
	// 2. The storage "constructor" should now return either `MemStorage` or `FileStorage`, depending on the passed configuration
	// 3. `MemStorage` and `FileStorage`  should implement the same interface. The controller should behave the same way
	//		no matter the actual implementation of the storage
	file *os.File
	mu   sync.Mutex
}

func NewMemStorage(cfg *config.ServerConfig) (*MemStorage, error) {
	file, err := os.OpenFile(cfg.StorePath, os.O_RDWR|os.O_CREATE, ownerFilePerm)
	if err != nil {
		return nil, fmt.Errorf("store failed to open file: %w", err)
	}
	fmt.Println("created")

	ms := &MemStorage{
		mu: sync.Mutex{},
		MetricsStorage: MetricsStorage{
			Counter: make(map[entities.MetricName]entities.Counter),
			Gauge:   make(map[entities.MetricName]entities.Gauge),
		},
		cfg:          cfg,
		stopSaveChan: make(chan struct{}),
		file:         file,
	}

	err = ms.loadFromFile()
	if err != nil {
		return nil, fmt.Errorf("storage mem filed to load the file: %w", err)
	}

	if cfg.StoreInterval > 0 {
		go ms.periodicSave()
	}

	return ms, nil
}

func (ms *MemStorage) CreateCounterRecord(metric entities.Metrics) error {
	ms.mu.Lock()
	_, ok := ms.Counter[entities.MetricName(metric.ID)]
	if !ok {
		ms.Counter[entities.MetricName(metric.ID)] = entities.Counter(*metric.Delta)
	} else {
		ms.Counter[entities.MetricName(metric.ID)] += entities.Counter(*metric.Delta)
	}
	ms.mu.Unlock()

	if ms.cfg.StoreInterval == 0 {
		err := ms.saveToFile()
		if err != nil {
			return fmt.Errorf("server service failed to save data to file %w", err)
		}
	}

	return nil
}

func (ms *MemStorage) CreateGaugeRecord(metric entities.Metrics) error {
	ms.mu.Lock()
	ms.Gauge[entities.MetricName(metric.ID)] = entities.Gauge(*metric.Value)
	ms.mu.Unlock()

	if ms.cfg.StoreInterval == 0 {
		return ms.saveToFile()
	}

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

func (ms *MemStorage) loadFromFile() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	fileInfo, err := ms.file.Stat()
	if err != nil {
		return fmt.Errorf("server service failed to get file info %w", err)
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(ms.file)

	data := NewMetricsStorage()

	err = decoder.Decode(&data)

	if err != nil {
		return fmt.Errorf("storage mem failed to json parse file content %w", err)
	}

	ms.Counter = data.Counter
	ms.Gauge = data.Gauge

	return nil
}

func (ms *MemStorage) saveToFile() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	err := ms.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("storage mem failed to truncate the file: %w", err)
	}
	_, err = ms.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("storage mem filed to reset the file: %w", err)
	}

	encoder := json.NewEncoder(ms.file)
	data := struct {
		Counter map[entities.MetricName]entities.Counter `json:"Counter"`
		Gauge   map[entities.MetricName]entities.Gauge   `json:"Gauge"`
	}{
		Counter: ms.Counter,
		Gauge:   ms.Gauge,
	}

	if err = encoder.Encode(&data); err != nil {
		return fmt.Errorf("storage mem filed to encode: %w", err)
	}
	return nil
}

func (ms *MemStorage) periodicSave() {
	ticker := time.NewTicker(time.Second * time.Duration(ms.cfg.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := ms.saveToFile(); err != nil {
				ms.cfg.Log.ErrorContext(context.Background(), "error saving the file")
			}
			ms.cfg.Log.DebugContext(context.Background(), "saved storage state")
		case <-ms.stopSaveChan:
			return
		}
	}
}

func (ms *MemStorage) Close() error {
	close(ms.stopSaveChan)
	err := ms.saveToFile()
	if err != nil {
		return fmt.Errorf("storage mem failed to save the file: %w", err)
	}

	err = ms.file.Close()
	if err != nil {
		return fmt.Errorf("storage mem failed to close the file %w", err)
	}
	return nil
}
