package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mihailtudos/metrickit/internal/compressor"
	"github.com/mihailtudos/metrickit/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressResponseWriter(t *testing.T) {
	tests := []struct {
		name           string
		encoding       string
		acceptEncoding string
		contentType    string
		responseCode   int
		responseBody   []byte
		requestBody    []byte
		compressed     bool
	}{
		{
			name:           "request doesn't accept or send compressed data",
			encoding:       "",
			acceptEncoding: "",
			responseCode:   http.StatusOK,
			responseBody:   nil,
			requestBody:    nil,
			compressed:     false,
			contentType:    "text/plain",
		},
		{
			name:           "can handle compressed request data",
			encoding:       "gzip",
			acceptEncoding: "",
			responseCode:   http.StatusOK,
			responseBody:   nil,
			requestBody:    bytes.Repeat([]byte("this is some repetitive content "), 100),
			compressed:     true,
			contentType:    "text/plain",
		},
		{
			name:           "returns bad request if request body is not compressed",
			encoding:       "gzip",
			acceptEncoding: "",
			responseCode:   http.StatusBadRequest,
			responseBody:   nil,
			requestBody:    bytes.Repeat([]byte("this is some repetitive content "), 100),
			compressed:     false,
			contentType:    "text/plain",
		},
		{
			name:           "accept compressed request data",
			encoding:       "",
			acceptEncoding: "gzip",
			responseCode:   http.StatusOK,
			responseBody:   bytes.Repeat([]byte("<h1>this is some repetitive content<h1>"), 100),
			requestBody:    nil,
			compressed:     true,
			contentType:    "text/html",
		},
		{
			name:           "accept compressed but sent uncomporessable content-type",
			encoding:       "",
			acceptEncoding: "gzip",
			responseCode:   http.StatusOK,
			responseBody:   bytes.Repeat([]byte("<h1>this is some repetitive content<h1>"), 100),
			requestBody:    nil,
			compressed:     false,
			contentType:    "text/text",
		},
		{
			name:           "sends compressed request data and receives compressed response",
			encoding:       "gzip",
			acceptEncoding: "gzip",
			responseCode:   http.StatusOK,
			responseBody:   bytes.Repeat([]byte("<h1>this is some repetitive content<h1>"), 100),
			requestBody:    bytes.Repeat([]byte("<h1>this is some repetitive content<h1>"), 100),
			compressed:     true,
			contentType:    "text/html",
		},
	}

	for _, tt := range tests {
		logBuf := new(bytes.Buffer)
		log, err := logger.NewLogger(logBuf, "info")
		require.NoError(t, err)
		var requestBody io.Reader

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", tt.contentType)
			w.WriteHeader(tt.responseCode)

			if _, err := w.Write(tt.responseBody); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

		rr := httptest.NewRecorder()

		if tt.encoding == "gzip" && tt.compressed {
			cps := compressor.NewCompressor(log)
			compressedData, err := cps.Compress(tt.requestBody)
			require.NoError(t, err)
			requestBody = bytes.NewReader(compressedData)
		} else {
			requestBody = bytes.NewReader(tt.requestBody)
		}

		r := httptest.NewRequest(http.MethodGet, "/test", requestBody)
		r.Header.Set("Content-Encoding", tt.encoding)
		r.Header.Set("Content-Type", tt.contentType)
		r.Header.Set("Accept-Encoding", tt.acceptEncoding)

		withCompressedResponseHandler := WithCompressedResponse(log)(handler)
		withCompressedResponseHandler.ServeHTTP(rr, r)

		assert.Equal(t, tt.responseCode, rr.Code)
		if tt.compressed {
			assert.Equal(t, tt.acceptEncoding, rr.Header().Get("Content-Encoding"))
			if tt.responseBody != nil && tt.requestBody != nil {
				cps := compressor.NewCompressor(log)
				decompressedBody, err := cps.Decompress(rr.Body)
				require.NoError(t, err)

				assert.Equal(t, string(tt.responseBody), string(decompressedBody))
			}
		} else {
			assert.Equal(t, "", rr.Header().Get("Content-Encoding"))
		}
	}
}
