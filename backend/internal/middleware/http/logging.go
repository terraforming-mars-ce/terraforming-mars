package httpmiddleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// newResponseWriter creates a new response writer wrapper
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default status
	}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("wrapped ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

// LoggingMiddleware logs HTTP requests using Zap
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer
		wrapped := newResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate request duration
		duration := time.Since(start)

		// Build log fields
		fields := []zap.Field{
			zap.Int("status", wrapped.statusCode),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Duration("duration", duration),
		}

		// Add query parameters if present
		if r.URL.RawQuery != "" {
			fields = append(fields, zap.String("query", r.URL.RawQuery))
		}

		// Log based on status code
		msg := "HTTP Request"

		if wrapped.statusCode >= 500 {
			logger.Get().Error(msg, fields...)
		} else if wrapped.statusCode >= 400 {
			logger.Get().Warn(msg, fields...)
		} else {
			logger.Get().Debug(msg, fields...)
		}
	})
}
