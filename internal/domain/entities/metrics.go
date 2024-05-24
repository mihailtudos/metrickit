package entities

import "errors"

var ErrJSONMarshal = errors.New("failed to marshal to JSON")

type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // value for metrics of type counter
	Value *float64 `json:"value,omitempty"` // value for metrics of type gauge
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // metric type gauge | counter
}
