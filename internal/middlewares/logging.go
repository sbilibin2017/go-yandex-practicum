package middlewares

import (
	"net/http"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware logs incoming HTTP requests and responses including
// URI, method, duration of the request processing, response status code,
// and response size in bytes.
//
// It wraps the http.ResponseWriter to capture response status and size.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default status code
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Log.Info("Request info",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
		)

		logger.Log.Info("Response info",
			zap.Int("status", rw.statusCode),
			zap.Int("size", rw.size),
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures
// the HTTP status code and size of the response body.
type responseWriter struct {
	http.ResponseWriter
	statusCode int // HTTP status code of the response
	size       int // Number of bytes written in the response body
}

// WriteHeader captures the status code and calls the underlying
// ResponseWriter's WriteHeader method.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the connection as part of an HTTP reply and
// captures the number of bytes written to keep track of response size.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}
