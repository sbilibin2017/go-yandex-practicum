package middlewares

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgconn"
)

// RetryConfig holds configuration options for RetryMiddleware.
//
// Delays specifies the duration to wait before each retry attempt.
type RetryConfig struct {
	Delays []time.Duration
}

// RetryOption defines a functional option for configuring RetryMiddleware.
type RetryOption func(*RetryConfig)

// NewRetryConfig creates a new RetryConfig applying the given options.
//
// By default, it uses the retry delays: 0, 1s, 3s, and 5s.
func NewRetryConfig(opts ...RetryOption) (*RetryConfig, error) {
	cfg := &RetryConfig{
		Delays: []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg, nil
}

// WithRetryDelays sets custom retry delays for RetryMiddleware.
//
// Example usage:
//
//	RetryMiddleware(WithRetryDelays(0, 2*time.Second, 5*time.Second))
func WithRetryDelays(delays ...time.Duration) RetryOption {
	return func(cfg *RetryConfig) {
		cfg.Delays = delays
	}
}

// RetryMiddleware returns an HTTP middleware that retries requests upon retriable errors.
//
// The middleware executes the wrapped handler up to len(cfg.Delays) times, waiting
// for the configured delay durations before each retry (except the first).
//
// A retriable error is detected using the isRetriableError helper function, which
// currently considers certain PostgreSQL errors and transient filesystem errors as retriable.
//
// If all retry attempts fail, it responds with HTTP 503 Service Unavailable.
//
// Custom retry delays can be provided via options such as WithRetryDelays.
func RetryMiddleware(opts ...RetryOption) (func(http.Handler) http.Handler, error) {
	cfg, err := NewRetryConfig(opts...)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			maxAttempts := len(cfg.Delays)

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				if attempt > 1 {
					time.Sleep(cfg.Delays[attempt-1])
				}

				brw := newBufferedResponseWriter()

				next.ServeHTTP(brw, r)

				if brw.err == nil {
					brw.flushTo(w)
					return
				}

				if !isRetriableError(brw.err) {
					brw.flushTo(w)
					return
				}
			}

			w.WriteHeader(http.StatusServiceUnavailable)
		})
	}, nil
}

// bufferedResponseWriter buffers the HTTP response to support retries.
//
// It captures headers, status code, and response body without writing to the client
// immediately, enabling the middleware to decide whether to retry or flush the response.
type bufferedResponseWriter struct {
	headers    http.Header
	statusCode int
	buf        bytes.Buffer
	err        error
}

// newBufferedResponseWriter creates a new bufferedResponseWriter with
// default status code 200 OK.
func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

// Header returns the buffered HTTP headers.
func (b *bufferedResponseWriter) Header() http.Header {
	return b.headers
}

// WriteHeader buffers the HTTP status code for the response.
func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

// Write buffers the response body data.
//
// If a write error has already occurred, subsequent writes return the same error.
func (b *bufferedResponseWriter) Write(data []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}

	n, err := b.buf.Write(data)
	if err != nil {
		b.err = err
	}
	return n, err
}

// flushTo writes the buffered headers, status code, and body to the provided ResponseWriter.
func (b *bufferedResponseWriter) flushTo(w http.ResponseWriter) {
	for k, vv := range b.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(b.statusCode)
	w.Write(b.buf.Bytes())
}

// isRetriableError determines if an error is considered retriable.
//
// It treats PostgreSQL connection errors (error codes starting with "08")
// and certain transient filesystem errors (EAGAIN, EWOULDBLOCK) as retriable.
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		if strings.HasPrefix(pgErr.Code, "08") {
			return true
		}
	}

	if pathErr, ok := err.(*os.PathError); ok {
		if pathErr.Err == syscall.EAGAIN || pathErr.Err == syscall.EWOULDBLOCK {
			return true
		}
	}

	return false
}
