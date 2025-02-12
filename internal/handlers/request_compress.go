// Package handlers provides HTTP handlers and middleware for the application.
package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mihailtudos/metrickit/internal/compressor"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// compressibleContentTypes holds the list of content types that can be compressed.
var compressibleContentTypes = map[string]struct{}{
	"application/json": {},
	"text/html":        {},
}

// compressResponseWriter is a custom http.ResponseWriter that writes compressed responses.
type compressResponseWriter struct {
	http.ResponseWriter              // The original ResponseWriter
	gzipWriter          *gzip.Writer // The gzip writer for compressing the response
	compressible        bool         // Flag indicating if the response is compressible
	wroteHeader         bool         // Flag indicating if the header has been written
}

// Write writes the data to the gzip writer if the response is compressible,
// or directly to the ResponseWriter otherwise. It ensures that the header
// is written before any body content.
func (crw *compressResponseWriter) Write(p []byte) (int, error) {
	if !crw.wroteHeader {
		crw.WriteHeader(http.StatusOK)
	}
	wb, err := crw.writer().Write(p)
	if err != nil {
		return 0, fmt.Errorf("failed to write content: %w", err)
	}
	return wb, nil
}

// writer returns the appropriate writer (gzip.Writer or ResponseWriter)
// based on whether the response is compressible.
func (crw *compressResponseWriter) writer() io.Writer {
	if crw.compressible {
		return crw.gzipWriter
	}
	return crw.ResponseWriter
}

// WriteHeader writes the HTTP status code and sets the Content-Encoding
// and Vary headers if the response is compressible.
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
		crw.Header().Del("Content-Length")                  // Remove Content-Length header
		crw.gzipWriter = gzip.NewWriter(crw.ResponseWriter) // Create new gzip writer
	}

	crw.ResponseWriter.WriteHeader(code) // Write the header to the original ResponseWriter
}

// isCompressible checks if the response's Content-Type is in the list of
// compressible content types.
func (crw *compressResponseWriter) isCompressible() bool {
	contentType := crw.Header().Get(helpers.ContentType)
	if strings.Contains(contentType, ";") {
		contentType = strings.Split(contentType, ";")[0] // Remove parameters from Content-Type
	}

	_, ok := compressibleContentTypes[contentType]
	return ok
}

// WithCompressedResponse is a middleware that compresses HTTP responses
// using Gzip if the client supports it and the content is compressible.
// It also handles decompression of incoming requests that are compressed.
func WithCompressedResponse(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decompress request body if it is gzip encoded
			if r.Header.Get("Content-Encoding") == "gzip" {
				c := compressor.NewCompressor(logger) // Create a new compressor instance
				decompressedBody, err := c.Decompress(r.Body)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(decompressedBody)) // Set the decompressed body
				r.Header.Del("Content-Encoding")                         // Remove Content-Encoding header
			}

			// If client does not accept gzip, serve without compression
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			crw := &compressResponseWriter{
				ResponseWriter: w, // Wrap the original ResponseWriter
			}

			next.ServeHTTP(crw, r) // Call the next handler
			// Close gzip writer if compression was performed
			if crw.compressible && crw.gzipWriter != nil {
				defer func() {
					if err := crw.gzipWriter.Close(); err != nil {
						logger.ErrorContext(r.Context(), "failed to close the gzip writer", helpers.ErrAttr(err))
					}
				}()
			}
		})
	}
}
