package handlers

import (
	"log/slog"
	"net"
	"net/http"
)

func WithRequestIPValidator(trustedIP *net.IPNet, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Real-IP")

			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}

			if ip != "" {
				IP := net.ParseIP(ip)
				if IP != nil && trustedIP.Contains(IP) {
					next.ServeHTTP(w, r)
					return
				} else {
					logger.ErrorContext(r.Context(),
						"request from untrusted IP",
						slog.String("ip", ip))
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
