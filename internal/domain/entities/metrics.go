// Package entities defines the data structures used for metrics in the metrics service.
package entities

// Metrics represents a metric entity with its associated properties.
// It can hold either a delta value for counter metrics or a value for gauge metrics.
// The `ID` field specifies the name of the metric, and the `MType` field indicates the type of metric.
//
// Fields:
// - Delta: An optional pointer to an int64 that holds the value for metrics of type counter.
// - Value: An optional pointer to a float64 that holds the value for metrics of type gauge.
// - ID: A string that uniquely identifies the metric. This field is required.
// - MType: A string that specifies the type of metric, which can be either "gauge" or "counter".
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // Value for metrics of type counter
	Value *float64 `json:"value,omitempty"` // Value for metrics of type gauge
	ID    string   `json:"id"`              // Metric name
	MType string   `json:"type"`            // Metric type: "gauge" | "counter"
}
