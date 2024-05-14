package repositories

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
)

type GaugeMemRepository struct {
	store *storage.MemStorage
}

func NewGaugeMemRepository(memStorage *storage.MemStorage) *GaugeMemRepository {
	return &GaugeMemRepository{store: memStorage}
}

func (gmr *GaugeMemRepository) Create(key string, gauge entities.Gauge) error {
	gmr.store.Mu.Lock()
	defer gmr.store.Mu.Unlock()
	gmr.store.Gauge[key] = gauge

	return nil
}

func (gmr *GaugeMemRepository) Get(key string) (entities.Gauge, bool) {
	gmr.store.Mu.Lock()
	defer gmr.store.Mu.Unlock()

	if gmr.store == nil || gmr.store.Gauge == nil {
		return entities.Gauge(0), false
	}

	v, ok := gmr.store.Gauge[key]
	return v, ok
}

func (gmr *GaugeMemRepository) GetAll() map[string]entities.Gauge {
	gmr.store.Mu.Lock()
	defer gmr.store.Mu.Unlock()

	if gmr.store == nil || gmr.store.Gauge == nil {
		return make(map[string]entities.Gauge)
	}

	copyMap := make(map[string]entities.Gauge)
	for k, v := range gmr.store.Gauge {
		copyMap[k] = v
	}

	return copyMap
}
