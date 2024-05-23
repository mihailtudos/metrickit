package entities

type MetricName string
type MetricType string

type MetricsCollection struct {
	GaugeMetrics   map[MetricName]Gauge
	CounterMetrics map[MetricName]Counter
}

func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		GaugeMetrics:   make(map[MetricName]Gauge),
		CounterMetrics: make(map[MetricName]Counter),
	}
}
