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
	gmr.store.Gauge[key] = gauge

	return nil
}
