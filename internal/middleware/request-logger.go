package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	logWriter struct {
		responseData *responseData
		http.ResponseWriter
	}
)

func (l *logWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write body: %w", err)
	}

	l.responseData.size += size
	return size, nil
}

func (l *logWriter) WriteHeader(status int) {
	l.responseData.status = status
	l.ResponseWriter.WriteHeader(status)
}

func RequestLogger(h http.Handler, logger *slog.Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method
		start := time.Now()
		proto := r.Proto

		respData := &responseData{status: http.StatusOK}
		lw := &logWriter{responseData: respData, ResponseWriter: w}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.InfoContext(r.Context(),
			"incoming "+proto,
			slog.String("uri", uri),
			slog.String("method", method),
			slog.String("duration", duration.String()),
			slog.Int("status", respData.status),
			slog.Int("size", respData.size),
		)
	}

	return http.HandlerFunc(fn)
}
