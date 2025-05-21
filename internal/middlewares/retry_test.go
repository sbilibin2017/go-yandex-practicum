package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestNewBufferedResponseWriter(t *testing.T) {
	bw := newBufferedResponseWriter()

	assert.NotNil(t, bw)
	assert.Equal(t, http.StatusOK, bw.statusCode)
	assert.NotNil(t, bw.headers)
	assert.Empty(t, bw.headers)
	assert.Zero(t, bw.buf.Len())
}

func TestBufferedResponseWriter_Header(t *testing.T) {
	bw := newBufferedResponseWriter()
	bw.Header().Set("X-Test", "value")
	assert.Equal(t, "value", bw.headers.Get("X-Test"))
}

func TestBufferedResponseWriter_WriteHeader(t *testing.T) {
	bw := newBufferedResponseWriter()
	bw.WriteHeader(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, bw.statusCode)
}

func TestBufferedResponseWriter_Write(t *testing.T) {
	bw := newBufferedResponseWriter()

	n, err := bw.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", bw.buf.String())
}

func TestBufferedResponseWriter_Write_WithError(t *testing.T) {
	bw := newBufferedResponseWriter()
	bw.err = errors.New("write failed")

	n, err := bw.Write([]byte("fail"))
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "write failed", err.Error())
	assert.Zero(t, bw.buf.Len())
}

func TestRetryMiddleware_NoError_ShouldFlushOnce(t *testing.T) {
	attempts := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		_, _ = w.Write([]byte("OK"))
	})

	retryHandler := RetryMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	retryHandler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	body := rec.Body.String()

	assert.Equal(t, 1, attempts, "Handler should be called only once")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "OK", body)
}

func TestBufferedResponseWriter_FlushTo(t *testing.T) {
	bw := newBufferedResponseWriter()
	bw.WriteHeader(http.StatusCreated)
	bw.Header().Set("X-Custom", "abc")
	bw.Write([]byte("response body"))

	rr := httptest.NewRecorder()
	bw.flushTo(rr)

	result := rr.Result()
	defer result.Body.Close()
	body := rr.Body.String()

	assert.Equal(t, http.StatusCreated, result.StatusCode)
	assert.Equal(t, "abc", result.Header.Get("X-Custom"))
	assert.Equal(t, "response body", body)
}

func TestRetryMiddleware_MaxRetriesExceeded(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bw, ok := w.(*bufferedResponseWriter); ok {
			bw.err = errors.New("write error")
		}
		_, _ = w.Write([]byte("fail"))
	})

	retryHandler := RetryMiddleware(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	retryHandler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Maximum retry attempts exceeded") {
		t.Errorf("expected error message in body, got %q", body)
	}
}

func TestIsRetriableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "pgconn retriable error",
			err:  &pgconn.PgError{Code: "08006"},
			want: true,
		},
		{
			name: "pgconn non-retriable error",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "os.PathError with EAGAIN",
			err: &os.PathError{
				Op:   "open",
				Path: "/some/path",
				Err:  syscall.EAGAIN,
			},
			want: true,
		},
		{
			name: "os.PathError with EWOULDBLOCK",
			err: &os.PathError{
				Op:   "open",
				Path: "/some/path",
				Err:  syscall.EWOULDBLOCK,
			},
			want: true,
		},
		{
			name: "os.PathError with different errno",
			err: &os.PathError{
				Op:   "open",
				Path: "/some/path",
				Err:  syscall.EPERM,
			},
			want: false,
		},
		{
			name: "unknown error",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetriableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
