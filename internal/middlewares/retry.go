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

func RetryMiddleware(next http.Handler) http.Handler {
	const maxAttempts = 4
	delays := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var brw *BufferedResponseWriter
		var err error

	retryLoop:
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			brw = NewBufferedResponseWriter()

			next.ServeHTTP(brw, r)

			if !isRetriableError(brw.err) {
				err = nil
				break
			}

			err = brw.err
			if attempt == maxAttempts {
				break
			}

			select {
			case <-time.After(delays[attempt-1]):
			case <-r.Context().Done():
				err = r.Context().Err()
				break retryLoop
			}
		}

		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		brw.flushTo(w)
	})
}

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

func (brw *BufferedResponseWriter) WriteHeader(statusCode int) {
	if brw.wroteHeader {
		return
	}
	brw.statusCode = statusCode
	brw.wroteHeader = true
}

func (brw *BufferedResponseWriter) Write(b []byte) (int, error) {
	if !brw.wroteHeader {
		brw.WriteHeader(http.StatusOK)
	}
	return brw.body.Write(b)
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
