package storage

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"sync"
)

type MetricsCollection struct {
	Mu         sync.Mutex
	Collection *entities.MetricsCollection
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		Mu:         sync.Mutex{},
		Collection: entities.NewMetricsCollection(),
	}
}
