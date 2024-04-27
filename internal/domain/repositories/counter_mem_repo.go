package repositories

import (
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
	_, ok := cmr.store.Counter[key]
	if !ok {
		cmr.store.Counter[key] = counter
	} else {
		cmr.store.Counter[key] += counter
	}

	return nil
}
