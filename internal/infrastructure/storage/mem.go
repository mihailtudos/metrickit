package storage

import (
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type MemStorage struct {
	Counter map[string]entities.Counter
	Gauge   map[string]entities.Gauge
	Mu      sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Mu:      sync.Mutex{},
		Counter: make(map[string]entities.Counter),
		Gauge:   make(map[string]entities.Gauge),
	}
}
