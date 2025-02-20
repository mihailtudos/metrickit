package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
	chiv5 "github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/mihailtudos/metrickit/internal/service/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDependencies(t *testing.T) server.Metrics {
	t.Helper()
	logger := slog.Default()
	store, err := storage.NewStorage(nil, logger, -1, ".")
	require.NoError(t, err)

	repos := repositories.NewRepository(store)
	service := server.NewMetricsService(repos, logger)
	return service
}

func setupTestRouter(t *testing.T, service server.Metrics) *chiv5.Mux {
	t.Helper()
	logger := slog.Default()

	mux := chiv5.NewRouter()
	serverHandlers := NewHandler(service, logger, nil, "", nil, nil)

	// Only use the request logger middleware for testing
	mux.Use(RequestLogger(logger))

	// Register routes directly without the validation middleware
	mux.Get("/value/{metricType}/{metricName}", serverHandlers.getMetricValue)
	mux.Get("/", serverHandlers.showMetrics(""))
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", serverHandlers.handleUploads)
	mux.Post("/update/", serverHandlers.handleJSONUploads)
	mux.Post("/updates/", serverHandlers.handleBatchUploads)
	mux.Post("/value/", serverHandlers.getJSONMetricValue)

	return mux
}

//nolint:exhaustive // Ignoring exhaustive check, only testing a subset of metrics
func TestServerHandler_showMetrics(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(m *mocks.MockMetrics)
		expectedStatus int
		expectedValues map[string]string
		templatePath   string
	}{
		{
			name: "successful metrics retrieval",
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().GetAll().Return(&storage.MetricsStorage{
					Gauge: map[entities.MetricName]entities.Gauge{
						"HeapAlloc": 1234.56,
					},
					Counter: map[entities.MetricName]entities.Counter{
						"TotalAlloc": 7890,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedValues: map[string]string{
				"HeapAlloc":  "1234.56",
				"TotalAlloc": "7890",
			},
		},
		{
			name: "service returns error",
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().GetAll().Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedValues: nil,
		},
		{
			name: "not existing template",
			setupMock: func(m *mocks.MockMetrics) {
				// Mock not called in this case
			},
			expectedStatus: http.StatusInternalServerError,
			expectedValues: nil,
			templatePath:   path.Join("tempaltes", "non-existent.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockMetrics(ctrl)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger, nil, "", nil, nil)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			w := httptest.NewRecorder()

			// Execute
			handler.showMetrics(tt.templatePath).ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
			}

			doc, err := goquery.NewDocumentFromReader(w.Body)
			assert.NoError(t, err)
			for metricName, expectedValue := range tt.expectedValues {
				value := doc.Find(fmt.Sprintf("strong:contains('%s')", metricName)).Parent().Text()
				assert.Contains(t, value, expectedValue)
			}
		})
	}
}

func TestServerHandler_getMetricValue(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		setupMock      func(m *mocks.MockMetrics)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "get valid counter metric",
			metricType: string(entities.CounterMetricName),
			metricName: "PollCount",
			setupMock: func(m *mocks.MockMetrics) {
				delta := int64(100)
				m.EXPECT().Get(entities.MetricName("PollCount"), entities.CounterMetricName).
					Return(entities.Metrics{
						ID:    "PollCount",
						MType: string(entities.CounterMetricName),
						Delta: &delta,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "100",
		},
		{
			name:       "get valid gauge metric",
			metricType: string(entities.GaugeMetricName),
			metricName: "HeapAlloc",
			setupMock: func(m *mocks.MockMetrics) {
				value := float64(123.45)
				m.EXPECT().Get(entities.MetricName("HeapAlloc"), entities.GaugeMetricName).
					Return(entities.Metrics{
						ID:    "HeapAlloc",
						MType: string(entities.GaugeMetricName),
						Value: &value,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
		},
		{
			name:       "metric not found",
			metricType: string(entities.GaugeMetricName),
			metricName: "NonExistent",
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("NonExistent"), entities.GaugeMetricName).
					Return(entities.Metrics{}, storage.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:       "invalid metric type",
			metricType: "invalid",
			metricName: "test",
			setupMock: func(m *mocks.MockMetrics) {
				// Mock not called for invalid type
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:       "service returns unexpected error",
			metricType: string(entities.GaugeMetricName),
			metricName: "HeapAlloc",
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("HeapAlloc"), entities.GaugeMetricName).
					Return(entities.Metrics{}, errors.New("unexpected database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockMetrics(ctrl)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger, nil, "", nil, nil)

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/value/%s/%s", tt.metricType, tt.metricName),
				http.NoBody)
			w := httptest.NewRecorder()

			// Create chi context with URL parameters
			rctx := chiv5.NewRouteContext()
			rctx.URLParams.Add("metricType", tt.metricType)
			rctx.URLParams.Add("metricName", tt.metricName)
			req = req.WithContext(context.WithValue(req.Context(), chiv5.RouteCtxKey, rctx))

			handler.getMetricValue(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
			}
		})
	}
}

func TestServerHandler_handleUploads(t *testing.T) {
	tests := []struct {
		name           string
		metricType     entities.MetricType
		metricName     entities.MetricName
		metricValue    string
		expectedStatus int
	}{
		{
			name:           "Valid gauge metric",
			metricType:     entities.GaugeMetricName,
			metricName:     entities.HeapAlloc,
			metricValue:    "123.12",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid counter metric",
			metricType:     entities.CounterMetricName,
			metricName:     entities.PollCount,
			metricValue:    "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid metric type",
			metricType:     "invalid",
			metricName:     "test",
			metricValue:    "123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid gauge value",
			metricType:     entities.GaugeMetricName,
			metricName:     "test_gauge",
			metricValue:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid counter value",
			metricType:     entities.CounterMetricName,
			metricName:     "test_counter",
			metricValue:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty metric value",
			metricType:     entities.GaugeMetricName,
			metricName:     "test",
			metricValue:    "",
			expectedStatus: http.StatusNotFound,
		},
	}
	serverService := setupDependencies(t)
	router := setupTestRouter(t, serverService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/update/%s/%s/%s", tt.metricType, tt.metricName, tt.metricValue)

			t.Logf("Making request to URL: %s", url)

			req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if status := rec.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				data, err := serverService.Get(tt.metricName, tt.metricType)
				require.NoError(t, err)
				t.Logf("data: %+v\n", data)

				if data.Value != nil {
					assert.Equal(t, tt.metricValue, fmt.Sprintf("%.2f", *data.Value))
				}

				if data.Delta != nil {
					assert.Equal(t, tt.metricValue,
						strconv.FormatInt(*data.Delta, 10))
				}
			}
		})
	}
}

//nolint:dupl // Intentionally similar test cases
func TestServerHandler_getJSONMetricValue(t *testing.T) {
	tests := []struct {
		name         string
		payload      entities.Metrics
		setupMock    func(m *mocks.MockMetrics)
		expectedCode int
		expectedBody *entities.Metrics
		testBody     io.Reader
	}{
		{
			name: "Valid Counter Metric",
			payload: entities.Metrics{
				ID:    "test_counter",
				MType: string(entities.CounterMetricName),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("test_counter"), entities.CounterMetricName).
					Return(entities.Metrics{
						ID:    "test_counter",
						MType: string(entities.CounterMetricName),
						Delta: int64Ptr(42),
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: &entities.Metrics{
				ID:    "test_counter",
				MType: string(entities.CounterMetricName),
				Delta: int64Ptr(42),
			},
		},
		{
			name: "Valid Gauge Metric",
			payload: entities.Metrics{
				ID:    "test_gauge",
				MType: string(entities.GaugeMetricName),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("test_gauge"), entities.GaugeMetricName).
					Return(entities.Metrics{
						ID:    "test_gauge",
						MType: string(entities.GaugeMetricName),
						Value: float64Ptr(123.45),
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: &entities.Metrics{
				ID:    "test_gauge",
				MType: string(entities.GaugeMetricName),
				Value: float64Ptr(123.45),
			},
		},
		{
			name: "Metric Not Found",
			payload: entities.Metrics{
				ID:    "non_existent",
				MType: string(entities.CounterMetricName),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("non_existent"), entities.CounterMetricName).
					Return(entities.Metrics{}, storage.ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Invalid Metric Type",
			payload: entities.Metrics{
				ID:    "test_invalid",
				MType: "invalid_type",
			},
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			payload: entities.Metrics{
				ID:    "test_error",
				MType: string(entities.CounterMetricName),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Get(entities.MetricName("test_error"), entities.CounterMetricName).
					Return(entities.Metrics{}, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "Body Read Error",
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusBadRequest,
			testBody:     &ErrorReader{Err: errors.New("forced read error")},
		},
		{
			name:         "Invalid JSON Body",
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusInternalServerError,
			testBody:     bytes.NewReader([]byte(`{invalid json}`)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockMetrics(ctrl)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger, nil, "", nil, nil)

			var req *http.Request
			if tt.testBody != nil {
				req = httptest.NewRequest(http.MethodPost, "/value/", tt.testBody)
			} else {
				payload, err := json.Marshal(tt.payload)
				require.NoError(t, err)
				req = httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(payload))
			}

			w := httptest.NewRecorder()

			handler.getJSONMetricValue(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != nil {
				var response entities.Metrics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, response)
			}
		})
	}
}

//nolint:dupl // Intentionally similar test cases
func TestServerHandler_handleJSONUploads(t *testing.T) {
	tests := []struct {
		name         string
		payload      entities.Metrics
		setupMock    func(m *mocks.MockMetrics)
		expectedCode int
		expectedBody *entities.Metrics
		testBody     io.Reader
	}{
		{
			name: "Valid Counter Metric",
			payload: entities.Metrics{
				ID:    "test_counter",
				MType: string(entities.CounterMetricName),
				Delta: int64Ptr(42),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Create(gomock.Any()).Return(nil)
				m.EXPECT().Get(entities.MetricName("test_counter"), entities.CounterMetricName).
					Return(entities.Metrics{
						ID:    "test_counter",
						MType: string(entities.CounterMetricName),
						Delta: int64Ptr(42),
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: &entities.Metrics{
				ID:    "test_counter",
				MType: string(entities.CounterMetricName),
				Delta: int64Ptr(42),
			},
		},
		{
			name:         "Invalid JSON Body",
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusBadRequest,
			testBody:     bytes.NewReader([]byte(`{invalid json}`)),
		},
		{
			name: "Service Create Error",
			payload: entities.Metrics{
				ID:    "test_error",
				MType: string(entities.GaugeMetricName),
				Value: float64Ptr(123.45),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Create(gomock.Any()).Return(errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "GetResponseMetric Error",
			payload: entities.Metrics{
				ID:    "test_error",
				MType: string(entities.CounterMetricName),
				Delta: int64Ptr(42),
			},
			setupMock: func(m *mocks.MockMetrics) {
				m.EXPECT().Create(gomock.Any()).Return(nil)
				m.EXPECT().Get(entities.MetricName("test_error"), entities.CounterMetricName).
					Return(entities.Metrics{}, errors.New("get error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid Metric Type",
			payload: entities.Metrics{
				ID:    "test_invalid",
				MType: "invalid_type",
				Value: float64Ptr(1.0),
			},
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Body Read Error",
			setupMock:    func(m *mocks.MockMetrics) {},
			expectedCode: http.StatusBadRequest,
			testBody:     &ErrorReader{Err: errors.New("forced read error")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockMetrics(ctrl)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger, nil, "", nil, nil)

			var req *http.Request
			if tt.testBody != nil {
				req = httptest.NewRequest(http.MethodPost, "/update/", tt.testBody)
			} else {
				payload, err := json.Marshal(tt.payload)
				require.NoError(t, err)
				req = httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(payload))
			}

			w := httptest.NewRecorder()
			handler.handleJSONUploads(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != nil {
				var response entities.Metrics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, response)
			}
		})
	}
}

// ErrorReader is a custom io.Reader that always returns an error.
type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, r.Err
}
