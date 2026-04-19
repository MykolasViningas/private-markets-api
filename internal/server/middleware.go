package server

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			logger.Info("request received",
				"method", r.Method,
				"path", r.URL.Path,
			)

			// wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, status: 200}

			next.ServeHTTP(ww, r)

			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
