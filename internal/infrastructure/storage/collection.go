package storage

import (
	"sync"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type MetricsCollection struct {
	Collection *entities.MetricsCollection
	Mu         sync.Mutex
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		Mu:         sync.Mutex{},
		Collection: entities.NewMetricsCollection(),
	}
}
