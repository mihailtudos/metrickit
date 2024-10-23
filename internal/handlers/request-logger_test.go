package handlers

import (
	"bytes"
	"fmt"
	"github.com/mihailtudos/metrickit/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogWriter(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		method       string
		responseCode int
		responseBody string
	}{
		{
			name:         "successful request logs correct parameters",
			path:         "/test",
			method:       http.MethodGet,
			responseCode: http.StatusOK,
			responseBody: "OK",
		},
		{
			name:         "bad request logs correct parameters",
			path:         "/test",
			method:       http.MethodGet,
			responseCode: http.StatusBadRequest,
			responseBody: "OK",
		},
		{
			name:         "server error request logs correct parameters",
			path:         "/test",
			method:       http.MethodPost,
			responseCode: http.StatusInternalServerError,
			responseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				_, err := w.Write([]byte(tt.responseBody))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			logBuf := new(bytes.Buffer)
			l, err := logger.NewLogger(logBuf, "info")
			require.NoError(t, err)

			loggedHandler := RequestLogger(l)(handler)
			r := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			loggedHandler.ServeHTTP(rr, r)
			assert.Equal(t, tt.responseCode, rr.Code)

			logOutput := logBuf.String()
			assert.Contains(t, logOutput, fmt.Sprintf(`"uri":"%s"`, tt.path))
			assert.Contains(t, logOutput, fmt.Sprintf(`"method":"%s"`, tt.method))
			assert.Contains(t, logOutput, fmt.Sprintf(`"status":%d`, tt.responseCode))
			assert.Contains(t, logOutput, `"duration":`)
			assert.Contains(t, logOutput, fmt.Sprintf(`"size":%d`, len(tt.responseBody)))
		})
	}
}
