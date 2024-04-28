package entities

type Counter int64

const (
	CounterMetricName            = "counter"
	PollCount         MetricName = "PollCount"
)

type CounterMetric struct {
	Name  MetricName
	Value Counter
}
