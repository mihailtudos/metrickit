// Package entities defines types and constants for representing metrics.
//
// It includes the definitions of Gauge and Counter types, along with
// constants for various metric names that are commonly used in
// performance monitoring and analysis.
package entities

// Gauge represents a metric that can increase or decrease, typically used
// to measure values at a specific point in time (e.g., temperature, memory usage).
type Gauge float64

// Counter represents a metric that only increases (or resets), used to count
// occurrences of events (e.g., number of requests, errors).
type Counter int64

// Predefined metric type constants for classification of metrics.
const (
	CounterMetricName MetricType = "counter" // Represents a metric that counts occurrences.
	GaugeMetricName   MetricType = "gauge"   // Represents a metric that measures values at a point in time.
)

// Predefined metric name constants for various metrics commonly used in Go.
// These constants provide meaningful names for tracking specific performance metrics.
const (
	Alloc           MetricName = "Alloc"           // Bytes allocated and still in use.
	BuckHashSys     MetricName = "BuckHashSys"     // Buck hash system memory.
	Frees           MetricName = "Frees"           // Total number of freed objects.
	GCCPUFraction   MetricName = "GCCPUFraction"   // Fraction of CPU time spent in garbage collection.
	GCSys           MetricName = "GCSys"           // Total memory allocated for garbage collection.
	HeapAlloc       MetricName = "HeapAlloc"       // Bytes allocated on the heap.
	HeapIdle        MetricName = "HeapIdle"        // Bytes in idle heap.
	HeapInuse       MetricName = "HeapInuse"       // Bytes in use in the heap.
	HeapObjects     MetricName = "HeapObjects"     // Number of allocated heap objects.
	HeapReleased    MetricName = "HeapReleased"    // Bytes released to the OS.
	HeapSys         MetricName = "HeapSys"         // Bytes allocated to the heap.
	LastGC          MetricName = "LastGC"          // Nanoseconds since last garbage collection.
	Lookups         MetricName = "Lookups"         // Number of pointer lookups.
	MCacheInuse     MetricName = "MCacheInuse"     // Bytes in use by the cache.
	MCacheSys       MetricName = "MCacheSys"       // Bytes allocated for the cache.
	MSpanInuse      MetricName = "MSpanInuse"      // Bytes in use by span.
	MSpanSys        MetricName = "MSpanSys"        // Bytes allocated for spans.
	Mallocs         MetricName = "Mallocs"         // Total number of memory allocations.
	NextGC          MetricName = "NextGC"          // Next scheduled garbage collection.
	NumForcedGC     MetricName = "NumForcedGC"     // Total number of forced garbage collections.
	NumGC           MetricName = "NumGC"           // Total number of garbage collections.
	OtherSys        MetricName = "OtherSys"        // Other system memory.
	PauseTotalNs    MetricName = "PauseTotalNs"    // Total pause time in nanoseconds.
	StackInuse      MetricName = "StackInuse"      // Bytes in use on the stack.
	StackSys        MetricName = "StackSys"        // Bytes allocated for the stack.
	Sys             MetricName = "Sys"             // Total bytes allocated by the program.
	TotalAlloc      MetricName = "TotalAlloc"      // Total bytes allocated by the program.
	RandomValue     MetricName = "RandomValue"     // Example metric name for a random value.
	PollCount       MetricName = "PollCount"       // Count of polling operations.
	TotalMemory     MetricName = "TotalMemory"     // Total memory allocated.
	FreeMemory      MetricName = "FreeMemory"      // Total free memory available.
	CPUutilization1 MetricName = "CPUutilization1" // Example metric for CPU utilization.
)

// GaugeMetric represents a gauge metric with its associated name and value.
type GaugeMetric struct {
	Name  MetricName // The name of the gauge metric.
	Value Gauge       // The value of the gauge metric.
}

// CounterMetric represents a counter metric with its associated name and value.
type CounterMetric struct {
	Name  MetricName // The name of the counter metric.
	Value Counter     // The value of the counter metric.
}
