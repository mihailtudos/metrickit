package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helperServerSetup(t *testing.T) ServerHandler {
	t.Helper()
	logger := slog.Default()
	memStore, err := storage.NewMemStorage(logger)
	require.NoError(t, err)

	repo := repositories.NewRepository(memStore)
	services := server.NewMetricsService(repo, logger)

	return ServerHandler{
		logger:      logger,
		TemplatesFs: templatesFs,
		db:          nil,
		secret:      "test",
		services:    services,
	}
}

func TestSingleUploadsHandler(t *testing.T) {
	sh := helperServerSetup(t)

	tests := []struct {
		name         string
		metricType   string
		metricName   string
		metricValue  string
		expectedCode int
	}{
		{
			name:         "Valid Counter Metric",
			metricType:   string(entities.CounterMetricName),
			metricName:   "requests",
			metricValue:  "10",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid Counter Metric Value",
			metricType:   string(entities.CounterMetricName),
			metricName:   "requests",
			metricValue:  "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Valid Gauge Metric",
			metricType:   string(entities.GaugeMetricName),
			metricName:   "cpu_usage",
			metricValue:  "0.75",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid Gauge Metric Value",
			metricType:   string(entities.GaugeMetricName),
			metricName:   "cpu_usage",
			metricValue:  "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Unsupported Metric Type",
			metricType:   "unsupported",
			metricName:   "some_metric",
			metricValue:  "100",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty Metric Value",
			metricType:   string(entities.CounterMetricName),
			metricName:   "requests",
			metricValue:  "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/updates/", http.NoBody)
			req = req.WithContext(context.Background())

			// Set URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", tt.metricType)
			rctx.URLParams.Add("metricName", tt.metricName)
			rctx.URLParams.Add("metricValue", tt.metricValue)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create a ResponseRecorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler
			sh.handleUploads(recorder, req)

			// Check the status code
			assert.Equal(t, tt.expectedCode, recorder.Code)
		})
	}
}

func TestHandleJSONUploads(t *testing.T) {
	sh := helperServerSetup(t)

	tests := []struct {
		name         string
		payload      entities.Metrics
		expectedCode int
	}{
		{
			name: "Valid Counter Metric",
			payload: entities.Metrics{
				MType: string(entities.CounterMetricName),
				ID:    "requests",
				Delta: int64Ptr(10),
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Valid Gauge Metric",
			payload: entities.Metrics{
				MType: string(entities.GaugeMetricName),
				ID:    "cpu_usage",
				Value: float64Ptr(0.75),
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid Metric Type",
			payload: entities.Metrics{
				MType: "unsupported",
				ID:    "some_metric",
				Delta: int64Ptr(100),
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Missing Metric ID",
			payload: entities.Metrics{
				MType: string(entities.CounterMetricName),
				Delta: int64Ptr(10),
			},
			expectedCode: http.StatusBadRequest,
		},
		// {
		// 	name: "Invalid JSON",
		// 	payload: entities.Metrics{
		// 		MType: "invalid",
		// 	},
		// 	expectedCode: http.StatusBadRequest,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/uploads/", bytes.NewReader(payload))
			req = req.WithContext(context.Background())

			recorder := httptest.NewRecorder()

			// Call the handler
			sh.handleJSONUploads(recorder, req)

			// Check the status code
			assert.Equal(t, tt.expectedCode, recorder.Code)

			// Optionally, check the response body for valid responses
			if tt.expectedCode == http.StatusOK {
				var responseMetric entities.Metrics
				err = json.Unmarshal(recorder.Body.Bytes(), &responseMetric)
				require.NoError(t, err)
				// Add any additional assertions on the responseMetric if necessary
			}
		})
	}
}

// Helper functions to create pointers for test cases.
func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestBatchUploadsHandler(t *testing.T) {
	sh := helperServerSetup(t)

	tests := []struct {
		name         string
		metrics      []entities.Metrics
		expectedCode int
		includeHash  bool
		correctHash  bool
		setupSecret  bool
		description  string
	}{
		{
			name: "Valid Batch Upload",
			metrics: []entities.Metrics{
				{ID: "requests", MType: string(entities.CounterMetricName), Delta: intPtr(10)},
				{ID: "cpu_usage", MType: string(entities.GaugeMetricName), Value: floatPtr(0.75)},
			},
			expectedCode: http.StatusOK,
			includeHash:  true,
			correctHash:  true,
			setupSecret:  true,
			description:  "Handles valid batch of metrics with correct hash and secret",
		},
		{
			name: "Invalid Metric Data",
			metrics: []entities.Metrics{
				{ID: "invalid", MType: "unsupported"},
			},
			expectedCode: http.StatusInternalServerError,
			includeHash:  false,
			correctHash:  false,
			setupSecret:  false,
			description:  "Returns error for unsupported metric type",
		},
		{
			name: "Hash Mismatch",
			metrics: []entities.Metrics{
				{ID: "requests", MType: string(entities.CounterMetricName), Delta: intPtr(5)},
			},
			expectedCode: http.StatusBadRequest,
			includeHash:  true,
			correctHash:  false,
			setupSecret:  true,
			description:  "Fails due to hash mismatch when secret is provided",
		},
		{
			name:         "Empty Request Body",
			metrics:      nil,
			expectedCode: http.StatusInternalServerError,
			includeHash:  false,
			correctHash:  false,
			setupSecret:  false,
			description:  "Handles an empty request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.metrics != nil {
				body, err = json.Marshal(tt.metrics)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/updates/", http.NoBody)
			req.Body = io.NopCloser(bytes.NewReader(body))

			// Set the hash header if needed
			if tt.includeHash {
				var hash string
				if tt.correctHash {
					hash = getHash(body, sh.secret)
				} else {
					hash = "incorrectHashValue"
				}
				req.Header.Set("HashSHA256", hash)
			}

			if tt.setupSecret {
				sh.secret = "test"
			} else {
				sh.secret = ""
			}

			// Create a ResponseRecorder to capture the response
			recorder := httptest.NewRecorder()

			// Call the handler
			sh.handleBatchUploads(recorder, req)

			// Check the status code
			assert.Equal(t, tt.expectedCode, recorder.Code, tt.description)
		})
	}
}

func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestGetHash(t *testing.T) {
	secret := "my_secret_key"
	data := []byte("test_data")

	// Calculate the expected hash manually
	expectedHash := hmac.New(sha256.New, []byte(secret))
	expectedHash.Write(data)
	expectedHashStr := hex.EncodeToString(expectedHash.Sum(nil))

	// Call getHash function
	hash := getHash(data, secret)

	// Compare the results
	assert.Equal(t, expectedHashStr, hash, "Expected hash does not match the actual hash")
}

func TestIsBodyValid(t *testing.T) {
	tests := []struct {
		data     []byte
		reqHash  string
		secret   string
		expected bool
	}{
		// Valid case
		{
			data:     []byte("test data"),
			reqHash:  getHash([]byte("test data"), "secret"),
			secret:   "secret",
			expected: true,
		},
		// Invalid hash
		{
			data:     []byte("test data"),
			reqHash:  "invalidhash",
			secret:   "secret",
			expected: false,
		},
		// Missing hash
		{
			data:     []byte("test data"),
			reqHash:  "",
			secret:   "secret",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.data), func(t *testing.T) {
			result := isBodyValid(tt.data, tt.reqHash, tt.secret)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsMetricNameAndValuePresent(t *testing.T) {
	tests := []struct {
		metric   entities.Metrics
		expected bool
	}{
		// Valid Counter Metric
		{
			metric: entities.Metrics{
				Delta: ptrInt64(10),
				ID:    "counter1",
				MType: string(entities.CounterMetricName),
			},
			expected: true,
		},
		// Invalid Counter Metric (missing Delta)
		{
			metric: entities.Metrics{
				ID:    "counter2",
				MType: string(entities.CounterMetricName),
			},
			expected: false,
		},
		// Valid Gauge Metric
		{
			metric: entities.Metrics{
				Value: ptrFloat64(10.5),
				ID:    "gauge1",
				MType: string(entities.GaugeMetricName),
			},
			expected: true,
		},
		// Invalid Gauge Metric (missing Value)
		{
			metric: entities.Metrics{
				ID:    "gauge2",
				MType: string(entities.GaugeMetricName),
			},
			expected: false,
		},
		// Invalid Metric Type
		{
			metric: entities.Metrics{
				ID:    "invalidMetric",
				MType: "invalid",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.metric.ID, func(t *testing.T) {
			result := isMetricNameAndValuePresent(tt.metric)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper functions to create pointers.
func ptrInt64(i int64) *int64 {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}
