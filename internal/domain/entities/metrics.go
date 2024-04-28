package entities

type MetricName string

type MetricsCollection struct {
	GaugeMetrics   []GaugeMetric
	CounterMetrics []CounterMetric
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		GaugeMetrics:   []GaugeMetric{},
		CounterMetrics: []CounterMetric{},
	}
}
