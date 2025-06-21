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

// RetryOption defines a functional option for configuring RetryMiddleware.
type RetryOption func(*retryMiddleware)

// retryMiddleware holds configuration and state for RetryMiddleware.
type retryMiddleware struct {
	delays []time.Duration
}

// WithRetryDelays sets custom retry delays for RetryMiddleware.
//
// Example:
//
//	RetryMiddleware(WithRetryDelays(0, 2*time.Second, 5*time.Second))
func WithRetryDelays(delays ...time.Duration) RetryOption {
	return func(mw *retryMiddleware) {
		mw.delays = delays
	}
}

// RetryMiddleware returns an HTTP middleware that retries requests upon retriable errors.
//
// It retries up to len(delays) times, waiting for configured delays before each retry (except first).
// If all retries fail, responds with HTTP 503 Service Unavailable.
func RetryMiddleware(opts ...RetryOption) (func(http.Handler) http.Handler, error) {
	mw := &retryMiddleware{
		delays: []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}, // default delays
	}

	for _, opt := range opts {
		opt(mw)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			maxAttempts := len(mw.delays)

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				if attempt > 1 {
					time.Sleep(mw.delays[attempt-1])
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
type bufferedResponseWriter struct {
	headers    http.Header
	statusCode int
	buf        bytes.Buffer
	err        error
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (b *bufferedResponseWriter) Header() http.Header {
	return b.headers
}

func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

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
