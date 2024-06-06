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
		return errors.New("failed to create the record: " + err.Error())
	}

	return nil
}

func (cmr *MetricsMemRepository) Get(key entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	record, err := cmr.store.GetRecord(key, mType)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Metrics{}, fmt.Errorf("item with key %s was not found: %w", key, err)
		}

		return entities.Metrics{}, errors.New("failed to get the item: " + err.Error())
	}

	return record, nil
}

func (cmr *MetricsMemRepository) GetAll() (*storage.MetricsStorage, error) {
	store, err := cmr.store.GetAllRecords()
	if err != nil {
		return nil, errors.New("failed to get the metrics: " + err.Error())
	}

	return store, nil
}

func (cmr *MetricsMemRepository) GetAllByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics,
	error) {
	metrics, err := cmr.store.GetAllRecordsByType(mType)
	if err != nil {
		return nil, errors.New("failed to get the metrics: " + err.Error())
	}

	return metrics, nil
}
