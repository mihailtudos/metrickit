package agent

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/mihailtudos/metrickit/proto/metrics"
	"google.golang.org/grpc"
	"log/slog"
	mrand "math/rand"
	"net"
	"net/http"
	"runtime"

	"github.com/mihailtudos/metrickit/internal/compressor"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/pkg/helpers"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const aesKeySize = 32

// MetricsCollectionService is a service for collecting and storing metrics.
type MetricsCollectionService struct {
	mRepo     repositories.MetricsCollectionRepository
	logger    *slog.Logger
	secret    *string
	publicKey *rsa.PublicKey
	gRPCConn  *grpc.ClientConn
}

// NewMetricsCollectionService creates a new MetricsCollectionService.
func NewMetricsCollectionService(
	repo repositories.MetricsCollectionRepository,
	logger *slog.Logger,
	secret *string,
	publicKey *rsa.PublicKey,
	gRPCConn *grpc.ClientConn) *MetricsCollectionService {
	return &MetricsCollectionService{
		mRepo:     repo,
		logger:    logger,
		secret:    secret,
		publicKey: publicKey,
		gRPCConn:  gRPCConn,
	}
}

// Collect collects metrics and stores them.
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

	gaugeMetrics := map[entities.MetricName]entities.Gauge{ //nolint:exhaustive // entities.PollCount is of type Counter
		entities.RandomValue:     entities.Gauge(mrand.Float64()),
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

// Send returns all metrics.
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

	if m.gRPCConn != nil {
		m.logger.DebugContext(ctx, "publishing metrics via gRPC")
		c := pb.NewMetricServiceClient(m.gRPCConn)

		grpcRequestMetrics := make([]*pb.Metric, 0, len(allMetrics))

		for _, metric := range allMetrics {
			mm := &pb.Metric{
				Id:    metric.ID,
				MType: metric.MType,
			}

			if mm.MType == string(entities.CounterMetricName) {
				mm.Delta = metric.Delta
			}
			if mm.MType == string(entities.GaugeMetricName) {
				mm.Value = metric.Value
			}

			grpcRequestMetrics = append(grpcRequestMetrics, mm)
		}

		res, errClient := c.CreateMetrics(ctx, &pb.CreateMetricsRequest{Metrics: grpcRequestMetrics})
		m.logger.DebugContext(ctx, fmt.Sprintf("response from gRPC server: %v", res))
		return errClient

	}

	err = m.publishMetric(ctx, url, "application/json", allMetrics, m.publicKey)
	if err != nil {
		m.logger.ErrorContext(ctx,
			"publishing the counter metrics failed: ",
			helpers.ErrAttr(err))
		return fmt.Errorf("sent metrics %w", err)
	}

	return nil
}

// ErrJSONMarshal is an error that occurs when the metrics cannot be marshaled to JSON.
var ErrJSONMarshal = errors.New("failed to marshal to JSON")

// publishMetric publishes the metrics to the server.
func (m *MetricsCollectionService) publishMetric(ctx context.Context, url,
	contentType string, metrics []entities.Metrics, publicKey *rsa.PublicKey) error {
	mJSONStruct, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed serialize the metrics: %w", ErrJSONMarshal)
	}

	// Encrypt the metrics using the public key
	var encryptedData []byte
	if publicKey != nil {
		aesKey := make([]byte, aesKeySize)
		if _, err = rand.Read(aesKey); err != nil {
			return fmt.Errorf("failed to generate AES key: %w", err)
		}

		// Encrypt the AES key with RSA
		encryptedKey, errEncr := rsa.EncryptPKCS1v15(rand.Reader, publicKey, aesKey)
		if errEncr != nil {
			return fmt.Errorf("failed to encrypt AES key: %w", errEncr)
		}

		// Create AES cipher
		block, errChiph := aes.NewCipher(aesKey)
		if errChiph != nil {
			return fmt.Errorf("failed to create AES cipher: %w", errChiph)
		}

		// Create GCM mode
		gcm, errGcm := cipher.NewGCM(block)
		if errGcm != nil {
			return fmt.Errorf("failed to create GCM: %w", errGcm)
		}

		// Create nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err = rand.Read(nonce); err != nil {
			return fmt.Errorf("failed to create nonce: %w", err)
		}

		// Encrypt data with AES-GCM
		aesEncrypted := gcm.Seal(nonce, nonce, mJSONStruct, nil)

		// Combine encrypted key and data
		// First 256 bytes will be RSA-encrypted AES key, rest is AES-encrypted data
		//nolint:gocritic // it avoids creating a new variable
		encryptedData = append(encryptedKey, aesEncrypted...)
	} else {
		encryptedData = mJSONStruct
	}

	c := compressor.NewCompressor(m.logger)

	gzipBuffer, err := c.Compress(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(gzipBuffer))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Encoding", "gzip")

	// Add header to indicate encryption
	if publicKey != nil {
		req.Header.Set("X-Encryption", "RSA-AES") // or just "encrypted"
	}

	if m.secret != nil {
		hash := hmac.New(sha256.New, []byte(*m.secret))
		hash.Write(mJSONStruct)
		hashedStr := hex.EncodeToString(hash.Sum(nil))

		req.Header.Set("HashSHA256", hashedStr)
		m.logger.DebugContext(ctx,
			"request body signed successfully")
	}

	// Set the X-Real-IP header with the client's IP address
	setIPHeader(req)

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

// setIPHeader sets the X-Real-IP header with the client's IP address.
func setIPHeader(req *http.Request) {
	localIP := "127.0.0.1" // default fallback
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					localIP = ipnet.IP.String()
					break
				}
			}
		}
	}

	// Add X-Real-IP header
	req.Header.Set("X-Real-IP", localIP)
}
