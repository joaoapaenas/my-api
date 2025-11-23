package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger is a middleware that logs the start and end of each request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get the Request ID from Chi middleware
		reqID := middleware.GetReqID(r.Context())

		// Create a wrapper to capture the status code and bytes written
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Log the request completion
			slog.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"request_id", reqID,
				"ip", r.RemoteAddr,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
