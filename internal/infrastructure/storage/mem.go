package storage

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"sync"
)

type MemStorage struct {
	Mu      sync.Mutex
	Counter map[string]entities.Counter
	Gauge   map[string]entities.Gauge
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Mu:      sync.Mutex{},
		Counter: make(map[string]entities.Counter),
		Gauge:   make(map[string]entities.Gauge),
	}
}
