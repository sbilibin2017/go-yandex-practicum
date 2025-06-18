// Package middlewares provides middleware components for HTTP servers.
package middlewares

import (
	"net/http"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware is an HTTP middleware that logs incoming requests and outgoing responses.
//
// For each request, it logs the HTTP method, URI, and processing duration.
// For each response, it logs the status code and the size of the response body.
//
// The logging is done using the zap logger from the internal logger package.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Log.Desugar().Info("Request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Duration("duration", duration),
		)

		logger.Log.Desugar().Info("Response",
			zap.Int("status", rw.statusCode),
			zap.Int("response_size", rw.writtenSize),
		)
	})
}

// responseWriter is a custom implementation of http.ResponseWriter
// that captures the response status code and the number of bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int // HTTP status code
	writtenSize int // total bytes written to the response body
}

// newResponseWriter creates and returns a new wrapped responseWriter instance.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code and delegates to the original ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the response body and updates the written size.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.writtenSize += n
	return n, err
}
