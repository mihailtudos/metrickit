package storage

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type MemStorage struct {
	Counter map[string]entities.Counter
	Gauge   map[string]entities.Gauge
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Counter: make(map[string]entities.Counter),
		Gauge:   make(map[string]entities.Gauge),
	}
}
