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
	cmr.store.Mu.Lock()
	defer cmr.store.Mu.Unlock()

	_, ok := cmr.store.Counter[key]
	if !ok {
		cmr.store.Counter[key] = counter
	} else {
		cmr.store.Counter[key] += counter
	}

	return nil
}

func (cmr *CounterMemRepository) Get(key string) (entities.Counter, bool) {
	cmr.store.Mu.Lock()
	defer cmr.store.Mu.Unlock()

	if cmr.store == nil || cmr.store.Counter == nil {
		return entities.Counter(0), false
	}
	val, ok := cmr.store.Counter[key]
	return val, ok
}

func (cmr *CounterMemRepository) GetAll() map[string]entities.Counter {
	cmr.store.Mu.Lock()
	defer cmr.store.Mu.Unlock()

	if cmr.store == nil || cmr.store.Counter == nil {
		return make(map[string]entities.Counter)
	}

	return cmr.store.Counter
}
