package repositories

import (
	"errors"
	"fmt"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type MetricsMemRepository struct {
	store storage.Storage
}

func NewMetricsMemRepository(store storage.Storage) *MetricsMemRepository {
	return &MetricsMemRepository{store: store}
}

func (cmr *MetricsMemRepository) Create(metric entities.Metrics) error {
	err := cmr.store.CreateRecord(metric)
	if err != nil {
		return fmt.Errorf("failed to create the record: %w", err)
	}

	return nil
}

func (cmr *MetricsMemRepository) Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	record, err := cmr.store.GetRecord(key, mType)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Metrics{}, fmt.Errorf("item with key %s was not found: %w", key, err)
		}

		return entities.Metrics{}, fmt.Errorf("failed to get the item: %w", err)
	}

	return record, nil
}

func (cmr *MetricsMemRepository) GetAll() (*storage.MetricsStorage, error) {
	store, err := cmr.store.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics: %w", err)
	}

	return store, nil
}

func (cmr *MetricsMemRepository) GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics,
	error) {
	metrics, err := cmr.store.GetAllRecordsByType(mType)
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics: %w", err)
	}

	return metrics, nil
}

func (cmr *MetricsMemRepository) StoreMetricsBatch(metrics []entities.Metrics) error {
	if err := cmr.store.StoreMetricsBatch(metrics); err != nil {
		return fmt.Errorf("mem storage error: %w", err)
	}

	return nil
}
