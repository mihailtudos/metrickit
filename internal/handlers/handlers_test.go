package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mihailtudos/metrickit/internal/service/agent"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUploads(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockService := agent.NewMockService()
	h := NewHandler(mockService, logger).InitHandlers()

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		url    string
		method string
		want
	}{
		{
			name:   "GET req method type not allowed",
			url:    "/update/counter/someMetric/232",
			method: http.MethodGet,
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
		{
			name:   "missing metric type from the url",
			url:    "/update/",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "passing invalid metric type",
			url:    "/update/unknown/testCounter/100",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name:   "missing metric name",
			url:    "/update/counter/",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "missing metric value",
			url:    "/update/counter/someMetric/",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "metric value is not of expected type",
			url:    "/update/counter/someMetric1/no-val",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "correct counter metric should return success",
			url:    "/update/counter/someMetric1/222",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "correct gauge metric should return success",
			url:    "/update/gauge/someMetric1/222",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, http.NoBody)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
