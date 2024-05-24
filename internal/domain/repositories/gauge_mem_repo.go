package repositories

import (
	"errors"
	"fmt"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type GaugeMemRepository struct {
	store *storage.MemStorage
}

func NewGaugeMemRepository(memStorage *storage.MemStorage) *GaugeMemRepository {
	return &GaugeMemRepository{store: memStorage}
}

func (gmr *GaugeMemRepository) Create(metric entities.Metrics) error {
	err := gmr.store.CreateGaugeRecord(metric)
	if err != nil {
		return errors.New("failed to create gauge record: " + err.Error())
	}
	return nil
}

func (gmr *GaugeMemRepository) Get(key entities.MetricName) (entities.Gauge, error) {
	item, err := gmr.store.GetGaugeRecord(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Gauge(0), fmt.Errorf("item with key %s was not found %w", key, err)
		}

		return entities.Gauge(0), errors.New("failed to get the item: " + err.Error())
	}

	return item, nil
}

func (gmr *GaugeMemRepository) GetAll() (map[entities.MetricName]entities.Gauge, error) {
	gauges, err := gmr.store.GetAllGaugeRecords()
	if err != nil {
		return nil, errors.New("failed to get the metrics: " + err.Error())
	}

	return gauges, nil
}
