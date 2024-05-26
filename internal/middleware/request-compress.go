package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mihailtudos/metrickit/pkg/compressor"
)

var compressibleContentTypes = map[string]struct{}{
	"application/json": {},
	"text/html":        {},
}

type compressResponseWriter struct {
	http.ResponseWriter
	gzipWriter   *gzip.Writer
	compressible bool
	wroteHeader  bool
}

func (crw *compressResponseWriter) Write(p []byte) (int, error) {
	if !crw.wroteHeader {
		crw.WriteHeader(http.StatusOK)
	}
	wb, err := crw.writer().Write(p)
	return wb, fmt.Errorf("failed to wrote body %w", err)
}

func (crw *compressResponseWriter) writer() io.Writer {
	if crw.compressible {
		return crw.gzipWriter
	}
	return crw.ResponseWriter
}

func (crw *compressResponseWriter) WriteHeader(code int) {
	if crw.wroteHeader {
		return
	}
	crw.wroteHeader = true

	// Set Content-Encoding and Vary headers if compressible
	if crw.isCompressible() {
		crw.compressible = true
		crw.Header().Set("Content-Encoding", "gzip")
		crw.Header().Add("Vary", "Accept-Encoding")
		crw.Header().Del("Content-Length")
		crw.gzipWriter = gzip.NewWriter(crw.ResponseWriter)
	}

	crw.ResponseWriter.WriteHeader(code)
}

func (crw *compressResponseWriter) isCompressible() bool {
	contentType := crw.Header().Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[0:idx]
	}

	_, ok := compressibleContentTypes[contentType]
	return ok
}

func WithCompressedResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			body, _ := io.ReadAll(r.Body)
			decompressedBody, err := compressor.Decompress(body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decompressedBody))
			r.Header.Del("Content-Encoding")
			fmt.Println("success")
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		crw := &compressResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(crw, r)

		if crw.compressible && crw.gzipWriter != nil {
			defer func() {
				if err := crw.gzipWriter.Close(); err != nil {
					fmt.Println("Error closing gzip writer:", err)
				}
			}()
		}
	})
}
