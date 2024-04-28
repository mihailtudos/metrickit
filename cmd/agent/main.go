package main

import (
	"fmt"
	"github.com/mihailtudos/metrickit/config"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func main() {
	agentCfg := config.NewAgentCfg(
		time.Second*2,
		time.Second*10,
		"http://localhost:8080",
	)
	metrics := &entities.MetricsCollection{}

	go collectMetrics(agentCfg.PollInterval, metrics)

	// http://<SERVER_ADDR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VAL>
	for {
		time.Sleep(agentCfg.ReportInterval)
		// publish counter type metrics
		for _, v := range metrics.CounterMetrics {
			err := publishMetric(entities.CounterMetricName, agentCfg.ServerAddr, v)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		// publish gauge type metrics
		for _, v := range metrics.GaugeMetrics {
			err := publishMetric(entities.GaugeMetricName, agentCfg.ServerAddr, v)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func collectMetrics(poolItl time.Duration, metrics *entities.MetricsCollection) {
	poolTicker := time.NewTicker(poolItl)

	currMetric := runtime.MemStats{}
	var counter entities.Counter

	for range poolTicker.C {
		*metrics = entities.MetricsCollection{}

		counter++
		runtime.ReadMemStats(&currMetric)
		// Counter Metrics
		metrics.CounterMetrics = append(metrics.CounterMetrics, entities.CounterMetric{Name: entities.PollCount, Value: counter})

		// Gauge Metrics
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.RandomValue, Value: entities.Gauge(rand.Float64())})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.Alloc, Value: entities.Gauge(currMetric.Alloc)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.BuckHashSys, Value: entities.Gauge(currMetric.BuckHashSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.Frees, Value: entities.Gauge(currMetric.Frees)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.GCCPUFraction, Value: entities.Gauge(currMetric.GCCPUFraction)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.GCSys, Value: entities.Gauge(currMetric.GCSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapAlloc, Value: entities.Gauge(currMetric.HeapAlloc)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapIdle, Value: entities.Gauge(currMetric.HeapIdle)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapInuse, Value: entities.Gauge(currMetric.HeapInuse)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapObjects, Value: entities.Gauge(currMetric.HeapObjects)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapReleased, Value: entities.Gauge(currMetric.HeapReleased)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.HeapSys, Value: entities.Gauge(currMetric.HeapSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.LastGC, Value: entities.Gauge(currMetric.LastGC)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.Lookups, Value: entities.Gauge(currMetric.Lookups)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.MCacheInuse, Value: entities.Gauge(currMetric.MCacheInuse)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.MCacheSys, Value: entities.Gauge(currMetric.MCacheSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.MSpanInuse, Value: entities.Gauge(currMetric.MSpanInuse)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.MSpanSys, Value: entities.Gauge(currMetric.MSpanSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.Mallocs, Value: entities.Gauge(currMetric.Mallocs)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.NextGC, Value: entities.Gauge(currMetric.NextGC)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.NumForcedGC, Value: entities.Gauge(currMetric.NumForcedGC)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.NumGC, Value: entities.Gauge(currMetric.NumGC)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.OtherSys, Value: entities.Gauge(currMetric.OtherSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.PauseTotalNs, Value: entities.Gauge(currMetric.PauseTotalNs)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.StackInuse, Value: entities.Gauge(currMetric.StackInuse)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.StackSys, Value: entities.Gauge(currMetric.StackSys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.Sys, Value: entities.Gauge(currMetric.Sys)})
		metrics.GaugeMetrics = append(metrics.GaugeMetrics, entities.GaugeMetric{Name: entities.TotalAlloc, Value: entities.Gauge(currMetric.TotalAlloc)})
	}
}

func publishMetric(mType, ServerAddr string, metric any) error {
	url := ""

	switch v := metric.(type) {
	case entities.CounterMetric:
		url = fmt.Sprintf("%s/update/%s/%s/%v", ServerAddr, entities.CounterMetricName, v.Name, v.Value)
	case entities.GaugeMetric:
		url = fmt.Sprintf("%s/update/%s/%s/%v", ServerAddr, entities.GaugeMetricName, v.Name, v.Value)
	}

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		fmt.Println("published successfully: ", url)
	}

	return nil
}
