package agent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"log/slog"
	"math/rand"
	"net/http"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/compressor"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/pkg/helpers"

	"github.com/shirou/gopsutil/v4/mem"
)

type MetricsCollectionService struct {
	mRepo  repositories.MetricsCollectionRepository
	logger *slog.Logger
	secret *string
}

func NewMetricsCollectionService(repo repositories.MetricsCollectionRepository,
	logger *slog.Logger,
	secret *string) *MetricsCollectionService {
	return &MetricsCollectionService{mRepo: repo, logger: logger, secret: secret}
}

func (m *MetricsCollectionService) Collect() error {
	m.logger.DebugContext(context.Background(), "collecting metrics...")

	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	cpuUtilization, err := cpu.Percent(0, false)
	if err != nil {
		return fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	gaugeMetrics := map[entities.MetricName]entities.Gauge{
		entities.RandomValue:     entities.Gauge(rand.Float64()),
		entities.Alloc:           entities.Gauge(stats.Alloc),
		entities.BuckHashSys:     entities.Gauge(stats.BuckHashSys),
		entities.Frees:           entities.Gauge(stats.Frees),
		entities.GCCPUFraction:   entities.Gauge(stats.GCCPUFraction),
		entities.GCSys:           entities.Gauge(stats.GCSys),
		entities.HeapAlloc:       entities.Gauge(stats.HeapAlloc),
		entities.HeapIdle:        entities.Gauge(stats.HeapIdle),
		entities.HeapInuse:       entities.Gauge(stats.HeapInuse),
		entities.HeapObjects:     entities.Gauge(stats.HeapObjects),
		entities.HeapReleased:    entities.Gauge(stats.HeapReleased),
		entities.HeapSys:         entities.Gauge(stats.HeapSys),
		entities.LastGC:          entities.Gauge(stats.LastGC),
		entities.Lookups:         entities.Gauge(stats.Lookups),
		entities.MCacheInuse:     entities.Gauge(stats.MCacheInuse),
		entities.MCacheSys:       entities.Gauge(stats.MCacheSys),
		entities.MSpanInuse:      entities.Gauge(stats.MSpanInuse),
		entities.MSpanSys:        entities.Gauge(stats.MSpanSys),
		entities.Mallocs:         entities.Gauge(stats.Mallocs),
		entities.NextGC:          entities.Gauge(stats.NextGC),
		entities.NumForcedGC:     entities.Gauge(stats.NumForcedGC),
		entities.NumGC:           entities.Gauge(stats.NumGC),
		entities.OtherSys:        entities.Gauge(stats.OtherSys),
		entities.PauseTotalNs:    entities.Gauge(stats.PauseTotalNs),
		entities.StackInuse:      entities.Gauge(stats.StackInuse),
		entities.StackSys:        entities.Gauge(stats.StackSys),
		entities.Sys:             entities.Gauge(stats.Sys),
		entities.TotalAlloc:      entities.Gauge(stats.TotalAlloc),
		entities.TotalMemory:     entities.Gauge(v.Total),
		entities.FreeMemory:      entities.Gauge(v.Free),
		entities.CPUutilization1: entities.Gauge(cpuUtilization[0]),
	}

	if err := m.mRepo.Store(gaugeMetrics); err != nil {
		return fmt.Errorf("failed to store the metrics: %w", err)
	}

	return nil
}

func (m *MetricsCollectionService) Send(serverAddr string) error {
	url := fmt.Sprintf("http://%s/updates/", serverAddr)
	ctx := context.Background()

	metrics, err := m.mRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to send the metrics: %w", err)
	}

	allMetrics := make([]entities.Metrics, 0, len(metrics.CounterMetrics)+len(metrics.CounterMetrics))

	m.logger.DebugContext(ctx, "publishing counter metrics")
	for k, v := range metrics.CounterMetrics {
		val := int64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	m.logger.DebugContext(ctx, "publishing gauge metrics")
	for k, v := range metrics.GaugeMetrics {
		val := float64(v)
		metric := entities.Metrics{
			ID:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: &val,
		}
		allMetrics = append(allMetrics, metric)
	}

	err = m.publishMetric(ctx, url, "application/json", allMetrics)
	if err != nil {
		m.logger.ErrorContext(ctx,
			"publishing the counter metrics failed: ",
			helpers.ErrAttr(err))
		return fmt.Errorf("sent metrics %w", err)
	}

	return nil
}

var ErrJSONMarshal = errors.New("failed to marshal to JSON")

func (m *MetricsCollectionService) publishMetric(ctx context.Context, url,
	contentType string, metrics []entities.Metrics) error {
	mJSONStruct, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed serialize the metrics: %w", ErrJSONMarshal)
	}

	c := compressor.NewCompressor(m.logger)

	gzipBuffer, err := c.Compress(mJSONStruct)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(gzipBuffer))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

	if m.secret != nil {
		hash := hmac.New(sha256.New, []byte(*m.secret))
		hash.Write(mJSONStruct)
		hashedStr := hex.EncodeToString(hash.Sum(nil))

		req.Header.Set("HashSHA256", hashedStr)
		m.logger.DebugContext(ctx,
			"request body signed successfully")
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post metric: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			m.logger.ErrorContext(ctx, "failed to close the body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to publish the metric " + res.Status)
	}

	m.logger.DebugContext(ctx, "published successfully", slog.String("metric", string(mJSONStruct)))
	return nil
}
