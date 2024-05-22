package repositories

import (
	"errors"
	"fmt"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type CounterMemRepository struct {
	store *storage.MemStorage
}

func NewCounterMemRepository(memStorage *storage.MemStorage) *CounterMemRepository {
	return &CounterMemRepository{store: memStorage}
}

func (cmr *CounterMemRepository) Create(key string, counter entities.Counter) error {
	err := cmr.store.CreateCounterRecord(key, counter)
	if err != nil {
		return errors.New("failed to create the record: " + err.Error())
	}

	return nil
}

func (cmr *CounterMemRepository) Get(key string) (entities.Counter, error) {
	item, err := cmr.store.GetCounterRecord(key)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Counter(0), fmt.Errorf("item with key %s was not found", key)
		}

		return entities.Counter(0), errors.New("failed to get the item: " + err.Error())
	}

	return item, nil
}

func (cmr *CounterMemRepository) GetAll() (map[string]entities.Counter, error) {
	counters, err := cmr.store.GetAllCounterRecords()
	if err != nil {
		return nil, errors.New("failed to get the metrics: " + err.Error())
	}

	return counters, nil
}
