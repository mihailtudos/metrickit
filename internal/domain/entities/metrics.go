package entities

import "sync"

type MetricName string

type MetricsCollection struct {
	Mu             sync.Mutex
	GaugeMetrics   []GaugeMetric
	CounterMetrics []CounterMetric
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		Mu:             sync.Mutex{},
		GaugeMetrics:   []GaugeMetric{},
		CounterMetrics: []CounterMetric{},
	}
}
