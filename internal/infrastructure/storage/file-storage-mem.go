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

var ownerFilePerm os.FileMode = 0o600

type FileStorage struct {
	stopSaveChan chan struct{}
	file         *os.File
	logger       *slog.Logger
	MemStorage
	storeInterval int
}

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
		return nil, fmt.Errorf("storage mem filed to load the file: %w", err)
	}

	if storeInterval > 0 {
		go fs.periodicSave()
	}

	return fs, nil
}

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

func (fs *FileStorage) Close(ctx context.Context) error {
	close(fs.stopSaveChan)
	err := fs.saveToFile()
	if err != nil {
		return fmt.Errorf("storage mem failed to save the file: %w", err)
	}

	err = fs.file.Close()
	if err != nil {
		return fmt.Errorf("storage mem failed to close the file %w", err)
	}
	return nil
}

func (fs *FileStorage) saveToFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := fs.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("storage mem failed to truncate the file: %w", err)
	}
	_, err = fs.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("storage mem filed to reset the file: %w", err)
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
		return fmt.Errorf("storage mem filed to encode: %w", err)
	}
	return nil
}

func (fs *FileStorage) loadFromFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fileInfo, err := fs.file.Stat()
	if err != nil {
		return fmt.Errorf("server service failed to get file info %w", err)
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(fs.file)

	data := NewMetricsStorage()

	err = decoder.Decode(&data)

	if err != nil {
		return fmt.Errorf("storage mem failed to json parse file content %w", err)
	}

	fs.Counter = data.Counter
	fs.Gauge = data.Gauge

	return nil
}

func (fs *FileStorage) CreateRecord(metrics entities.Metrics) error {
	if err := fs.MemStorage.CreateRecord(metrics); err != nil {
		return fmt.Errorf("file store: %w", err)
	}

	if fs.storeInterval == 0 {
		err := fs.saveToFile()
		if err != nil {
			return fmt.Errorf("server service failed to save data to file %w", err)
		}
	}

	return nil
}

func (fs *FileStorage) StoreMetricsBatch(metrics []entities.Metrics) error {
	if err := fs.MemStorage.StoreMetricsBatch(metrics); err != nil {
		return fmt.Errorf("file batch store: %w", err)
	}

	if fs.storeInterval == 0 {
		err := fs.saveToFile()
		if err != nil {
			return fmt.Errorf("server service failed to save data to file %w", err)
		}
	}

	return nil
}
