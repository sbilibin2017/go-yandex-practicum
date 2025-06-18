package middlewares

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgconn"
)

// BufferedResponseWriter buffers the response and captures an error.
type BufferedResponseWriter struct {
	headers     http.Header
	body        *bytes.Buffer
	statusCode  int
	wroteHeader bool
	err         error
}

func NewBufferedResponseWriter() *BufferedResponseWriter {
	return &BufferedResponseWriter{
		headers:    make(http.Header),
		body:       new(bytes.Buffer),
		statusCode: http.StatusOK,
	}
}

func (brw *BufferedResponseWriter) Header() http.Header {
	return brw.headers
}

func (brw *BufferedResponseWriter) Write(b []byte) (int, error) {
	return brw.body.Write(b)
}

func (brw *BufferedResponseWriter) WriteHeader(statusCode int) {
	if brw.wroteHeader {
		return
	}
	brw.statusCode = statusCode
	brw.wroteHeader = true
}

// SetError lets handler set an error that middleware will inspect for retries.
func (brw *BufferedResponseWriter) SetError(err error) {
	brw.err = err
}

// flushTo writes buffered headers, status and body to real ResponseWriter.
func (brw *BufferedResponseWriter) flushTo(w http.ResponseWriter) {
	for k, vv := range brw.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(brw.statusCode)
	w.Write(brw.body.Bytes())
}

func RetryMiddleware(next http.Handler) http.Handler {
	const maxAttempts = 4
	delays := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var brw *BufferedResponseWriter

		err := withRetry(r.Context(), maxAttempts, delays, func() error {
			brw = NewBufferedResponseWriter()

			next.ServeHTTP(brw, r)

			if isRetriableError(brw.err) {
				return brw.err
			}
			return nil
		})

		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		brw.flushTo(w)
	})
}

func withRetry(ctx context.Context, maxAttempts int, delays []time.Duration, fn func() error) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if attempt == maxAttempts {
			return err
		}
		if attempt-1 < len(delays) {
			select {
			case <-time.After(delays[attempt-1]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		return strings.HasPrefix(pgErr.Code, "08")
	}

	if pathErr, ok := err.(*os.PathError); ok {
		return pathErr.Err == syscall.EAGAIN || pathErr.Err == syscall.EWOULDBLOCK
	}

	return false
}
