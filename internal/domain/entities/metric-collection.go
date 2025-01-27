// Package entities defines data structures related to metrics handling.
//
// It provides types and functions to manage collections of metrics, including
// gauge and counter metrics.
package entities

// MetricName represents the name of a metric.
type MetricName string

// MetricType represents the type of a metric (e.g., gauge or counter).
type MetricType string

// MetricsCollection holds collections of gauge and counter metrics.
//
// It uses maps to store gauge and counter metrics, with MetricName as the key.
type MetricsCollection struct {
	GaugeMetrics   map[MetricName]Gauge   // GaugeMetrics maps metric names to their corresponding Gauge values.
	CounterMetrics map[MetricName]Counter // CounterMetrics maps metric names to their corresponding Counter values.
}

// NewMetricsCollection creates a new, empty instance of MetricsCollection.
//
// It initializes the maps for storing gauge and counter metrics.
//
// Returns:
//   - *MetricsCollection: A pointer to an empty MetricsCollection instance.
func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		GaugeMetrics:   make(map[MetricName]Gauge),
		CounterMetrics: make(map[MetricName]Counter),
	}
}
