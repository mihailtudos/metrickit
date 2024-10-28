// Package handlers provides HTTP handlers and middleware for the application.
package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type (
	// ResponseData holds information about the HTTP response status and size.
	responseData struct {
		status int // HTTP status code
		size   int // Size of the response body in bytes
	}

	// LogWriter wraps the standard http.ResponseWriter to capture response data.
	logWriter struct {
		responseData        *responseData // Reference to responseData to log information
		http.ResponseWriter               // The original ResponseWriter
	}
)

// Write writes the data to the response body and updates the size of the response.
// It returns the number of bytes written and any error encountered.
func (l *logWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b) // Write to the original ResponseWriter
	if err != nil {
		return 0, fmt.Errorf("failed to write body: %w", err) // Return error if write fails
	}

	l.responseData.size += size // Update the response size
	return size, nil
}

// WriteHeader writes the HTTP status code to the response and updates the status in responseData.
func (l *logWriter) WriteHeader(status int) {
	l.responseData.status = status       // Set the status in responseData
	l.ResponseWriter.WriteHeader(status) // Write the header to the original ResponseWriter
}

// RequestLogger is a middleware that logs incoming HTTP requests.
// It logs the request method, URI, duration of the request handling, response status, and size.
func RequestLogger(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Capture request details
			uri := r.RequestURI
			method := r.Method
			start := time.Now() // Record the start time
			proto := r.Proto    // HTTP protocol version

			respData := &responseData{status: http.StatusOK}            // Initialize responseData
			lw := &logWriter{responseData: respData, ResponseWriter: w} // Create a logWriter

			h.ServeHTTP(lw, r)            // Call the next handler in the chain
			duration := time.Since(start) // Calculate duration of request handling

			// Log the request details using the provided logger
			logger.InfoContext(r.Context(),
				"incoming "+proto,
				slog.String("uri", uri),
				slog.String("method", method),
				slog.String("duration", duration.String()),
				slog.Int("status", respData.status),
				slog.Int("size", respData.size),
			)
		}

		return http.HandlerFunc(fn) // Return the handler function
	}
}
