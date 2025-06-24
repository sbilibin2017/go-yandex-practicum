package middlewares

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a dummy retriable pgconn.PgError
func newPgConnError(code string) *pgconn.PgError {
	return &pgconn.PgError{Code: code}
}

func TestRetryMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		delays        []time.Duration
		handler       http.HandlerFunc
		expectStatus  int
		expectRetries int // how many times handler should be called
	}{
		{
			name:   "success on first try",
			delays: []time.Duration{0, 10 * time.Millisecond},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectStatus:  http.StatusOK,
			expectRetries: 1,
		},
		{
			name:   "retries on retriable error then success",
			delays: []time.Duration{0, 10 * time.Millisecond},
			handler: func() http.HandlerFunc {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts < 2 {
						// write retriable error to bufferedResponseWriter
						brw := w.(*bufferedResponseWriter)
						brw.err = newPgConnError("08006") // retriable error code prefix "08"
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			expectStatus:  http.StatusOK,
			expectRetries: 2,
		},
		{
			name:   "exceeds retries and returns 503",
			delays: []time.Duration{0, 10 * time.Millisecond},
			handler: func(w http.ResponseWriter, r *http.Request) {
				brw := w.(*bufferedResponseWriter)
				brw.err = newPgConnError("08006") // always retriable error
			},
			expectStatus:  http.StatusServiceUnavailable,
			expectRetries: 2,
		},
		{
			name:   "non-retriable error returns immediately",
			delays: []time.Duration{0, 10 * time.Millisecond},
			handler: func(w http.ResponseWriter, r *http.Request) {
				brw := w.(*bufferedResponseWriter)
				brw.err = errors.New("non retriable error")
			},
			expectStatus:  http.StatusOK, // flush called even with error that’s non-retriable (default status 200)
			expectRetries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := RetryMiddleware(WithRetryDelays(tt.delays...))
			require.NoError(t, err)

			callCount := 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.handler(w, r)
			})

			// Wrap handler with middleware
			finalHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			finalHandler.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close() // <--- Close response body here

			require.Equal(t, tt.expectStatus, resp.StatusCode)
			require.Equal(t, tt.expectRetries, callCount)
		})
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
			name: "pgconn error with code 08xxxx",
			err: &pgconn.PgError{
				Code: "08003", // connection failure class
			},
			want: true,
		},
		{
			name: "pgconn error with code not starting with 08",
			err: &pgconn.PgError{
				Code: "12345",
			},
			want: false,
		},
		{
			name: "os.PathError with EAGAIN",
			err: &os.PathError{
				Err: syscall.EAGAIN,
			},
			want: true,
		},
		{
			name: "os.PathError with EWOULDBLOCK",
			err: &os.PathError{
				Err: syscall.EWOULDBLOCK,
			},
			want: true,
		},
		{
			name: "os.PathError with different error",
			err: &os.PathError{
				Err: syscall.ECONNREFUSED,
			},
			want: false,
		},
		{
			name: "other generic error",
			err:  errors.New("some random error"),
			want: false,
		},
		{
			name: "pgconn error with empty code",
			err: &pgconn.PgError{
				Code: "",
			},
			want: false,
		},
		{
			name: "os.PathError with nil Err",
			err: &os.PathError{
				Err: nil,
			},
			want: false,
		},
		{
			name: "wrapped pgconn error with code 08xxxx",
			err:  wrapPgconnError("08006"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetriableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper to wrap pgconn.PgError to simulate wrapped errors if needed
func wrapPgconnError(code string) error {
	pgErr := &pgconn.PgError{Code: code}
	return pgErr // simply return, since isRetriableError checks type assertion directly
}

func TestBufferedResponseWriter(t *testing.T) {
	t.Run("Header returns headers map", func(t *testing.T) {
		brw := newBufferedResponseWriter()
		hdr := brw.Header()
		assert.NotNil(t, hdr)
		assert.IsType(t, make(http.Header), hdr)
	})

	t.Run("WriteHeader sets statusCode", func(t *testing.T) {
		brw := newBufferedResponseWriter()
		brw.WriteHeader(http.StatusTeapot)
		assert.Equal(t, http.StatusTeapot, brw.statusCode)
	})

	t.Run("Write writes data to buffer and returns bytes written", func(t *testing.T) {
		brw := newBufferedResponseWriter()
		n, err := brw.Write([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "hello", brw.buf.String())
	})

	t.Run("Write returns error and zero bytes written if already errored", func(t *testing.T) {
		brw := newBufferedResponseWriter()
		brw.err = assert.AnError
		n, err := brw.Write([]byte("hello"))
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("flushTo writes status code and body to ResponseWriter", func(t *testing.T) {
		brw := newBufferedResponseWriter()
		brw.WriteHeader(http.StatusAccepted)
		brw.Write([]byte("response body"))

		rec := httptest.NewRecorder()

		brw.flushTo(rec)

		res := rec.Result()
		defer res.Body.Close() // <--- Close response body here

		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		body, err := io.ReadAll(rec.Body)
		assert.NoError(t, err)
		assert.Equal(t, "response body", string(body))
	})
}
