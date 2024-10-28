// Package storage provides storage functionalities for metrics.
// It includes in-memory storage and file-based persistence for metrics data.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

// ownerFilePerm defines the permissions for the storage file.
var ownerFilePerm os.FileMode = 0o600

// FileStorage represents a storage backend that persists metrics to a file.
// It embeds MemStorage to utilize in-memory metrics handling and provides
// mechanisms for periodic saving of metrics to a file.
type FileStorage struct {
	stopSaveChan  chan struct{} // Channel for signaling when to stop saving
	file          *os.File      // File to persist metrics
	logger        *slog.Logger  // Logger for logging messages
	MemStorage                  // Embedded in-memory metrics storage
	storeInterval int           // Interval for periodic saving of metrics
}

// NewFileStorage creates a new instance of FileStorage. It opens the specified
// file for reading and writing, initializes the in-memory storage, and starts
// a periodic saving routine if the storeInterval is greater than zero.
func NewFileStorage(logger *slog.Logger, storeInterval int, storePath string) (*FileStorage, error) {
	file, err := os.OpenFile(storePath, os.O_RDWR|os.O_CREATE, ownerFilePerm)
	if err != nil {
		return nil, fmt.Errorf("store failed to open file: %w", err)
	}
	logger.DebugContext(context.Background(), "created file storage")

	fs := &FileStorage{
		logger: logger,
		MemStorage: MemStorage{
			mu: sync.Mutex{},
			MetricsStorage: MetricsStorage{
				Counter: make(map[entities.MetricName]entities.Counter),
				Gauge:   make(map[entities.MetricName]entities.Gauge),
			},
			logger: logger,
		},
		stopSaveChan:  make(chan struct{}),
		file:          file,
		storeInterval: storeInterval,
	}

	err = fs.loadFromFile()
	if err != nil {
		return nil, fmt.Errorf("storage mem failed to load the file: %w", err)
	}

	if storeInterval > 0 {
		go fs.periodicSave()
	}

	return fs, nil
}

// periodicSave periodically saves the in-memory metrics to the file based
// on the configured storeInterval. It runs in a separate goroutine.
func (fs *FileStorage) periodicSave() {
	ticker := time.NewTicker(time.Second * time.Duration(fs.storeInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := fs.saveToFile(); err != nil {
				fs.logger.ErrorContext(context.Background(), "error saving the file")
			}
			fs.logger.DebugContext(context.Background(), "saved storage state")
		case <-fs.stopSaveChan:
			return
		}
	}
}

// Close stops the periodic saving routine, saves any remaining metrics to the
// file, and closes the file.
func (fs *FileStorage) Close(ctx context.Context) error {
	close(fs.stopSaveChan)
	err := fs.saveToFile()
	if err != nil {
		return fmt.Errorf("storage mem failed to save the file: %w", err)
	}

	err = fs.file.Close()
	if err != nil {
		return fmt.Errorf("storage mem failed to close the file: %w", err)
	}
	return nil
}

// saveToFile saves the current state of the in-memory metrics to the file
// in JSON format, truncating the file first to ensure it's overwritten.
func (fs *FileStorage) saveToFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := fs.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("storage mem failed to truncate the file: %w", err)
	}
	_, err = fs.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("storage mem failed to reset the file: %w", err)
	}

	encoder := json.NewEncoder(fs.file)
	data := struct {
		Counter map[entities.MetricName]entities.Counter `json:"Counter"`
		Gauge   map[entities.MetricName]entities.Gauge   `json:"Gauge"`
	}{
		Counter: fs.Counter,
		Gauge:   fs.Gauge,
	}

	if err = encoder.Encode(&data); err != nil {
		return fmt.Errorf("storage mem failed to encode: %w", err)
	}
	return nil
}

// loadFromFile loads metrics from the file into the in-memory storage,
// populating the Counter and Gauge maps if data exists.
func (fs *FileStorage) loadFromFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fileInfo, err := fs.file.Stat()
	if err != nil {
		return fmt.Errorf("server service failed to get file info: %w", err)
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(fs.file)

	data := NewMetricsStorage()

	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("storage mem failed to json parse file content: %w", err)
	}

	fs.Counter = data.Counter
	fs.Gauge = data.Gauge

	return nil
}

// CreateRecord adds a new metric record to the in-memory storage and saves it
// to the file immediately if storeInterval is set to zero.
func (fs *FileStorage) CreateRecord(metrics entities.Metrics) error {
	if err := fs.MemStorage.CreateRecord(metrics); err != nil {
		return fmt.Errorf("file store: %w", err)
	}

	if fs.storeInterval == 0 {
		err := fs.saveToFile()
		if err != nil {
			return fmt.Errorf("server service failed to save data to file: %w", err)
		}
	}

	return nil
}

// StoreMetricsBatch adds multiple metric records to the in-memory storage
// and saves them to the file immediately if storeInterval is set to zero.
func (fs *FileStorage) StoreMetricsBatch(metrics []entities.Metrics) error {
	if err := fs.MemStorage.StoreMetricsBatch(metrics); err != nil {
		return fmt.Errorf("file batch store: %w", err)
	}

	if fs.storeInterval == 0 {
		err := fs.saveToFile()
		if err != nil {
			return fmt.Errorf("server service failed to save data to file: %w", err)
		}
	}

	return nil
}
