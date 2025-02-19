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
	// Value for metrics of type counter
	Delta *int64 `json:"delta,omitempty" protobuf:"varint,4,opt,name=delta,proto3,oneof"`
	// Value for metrics of type gauge
	Value *float64 `json:"value,omitempty" protobuf:"fixed64,3,opt,name=value,proto3,oneof"`
	// ID is the name of the metric
	ID string `json:"id,omitempty" protobuf:"bytes,1,opt,name=id,proto3"`
	// MType is the type of metric, which can be either "gauge" or "counter"
	MType string `json:"type,omitempty" protobuf:"bytes,2,opt,name=m_type,json=mType,proto3"`
}
