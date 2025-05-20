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

// bufferedResponseWriter буферизует ответ для возможности повторной попытки
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
	// Можно симулировать ошибку записи, например, через поле err
	if b.err != nil {
		return 0, b.err
	}
	return b.buf.Write(data)
}

func (b *bufferedResponseWriter) flushTo(w http.ResponseWriter) {
	for k, v := range b.headers {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(b.statusCode)
	w.Write(b.buf.Bytes())
}

func RetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const maxAttempts = 4
		delays := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if attempt > 1 {
				time.Sleep(delays[attempt-1])
			}

			brw := newBufferedResponseWriter()

			next.ServeHTTP(brw, r)

			if brw.err == nil && !isRetriableError(brw.err) {
				brw.flushTo(w)
				return
			}
		}

		http.Error(w, "Maximum retry attempts exceeded", http.StatusServiceUnavailable)
	})
}

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
