package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// WithBodyValidator is a middleware that checks the validity of request body.
func WithBodyValidator(secret string, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.ErrorContext(r.Context(),
					"failed to read request body: ",
					helpers.ErrAttr(err))
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if err := r.Body.Close(); err != nil {
				logger.ErrorContext(r.Context(),
					"failed to close request body: ",
					helpers.ErrAttr(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if secret != "" {
				if !isBodyValid(bodyBytes, r.Header.Get("HashSHA256"), secret) {
					logger.DebugContext(r.Context(),
						"request body failed integrity check")
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}

				logger.DebugContext(r.Context(),
					"request body passed integrity check")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isBodyValid verifies the integrity of the request body by comparing the provided hash with a computed hash.
// It returns true if the request hash is not empty and matches the computed hash using the provided secret.
func isBodyValid(data []byte, reqHash, secret string) bool {
	if reqHash == "" {
		return false
	}
	hashedStr := getHash(data, secret)
	return hashedStr == reqHash
}
