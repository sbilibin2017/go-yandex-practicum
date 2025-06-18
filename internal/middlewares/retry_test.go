package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

// testError simulates a retriable error
type testError struct{}

func (e *testError) Error() string { return "test retriable error" }

// Implement Temporary so isRetriableError can be adapted if needed (not used in your current code but often helpful)
func (e *testError) Temporary() bool { return true }

func TestRetryMiddleware_SuccessFirstTry(t *testing.T) {
	attempts := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	middleware := RetryMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <----- Close response body here
	body := w.Body.String()

	assert.Equal(t, 1, attempts, "Should call handler only once")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", body)
}

// Test the retry logic by calling withRetry directly, simulating errors
func TestWithRetry_SucceedsAfterRetries(t *testing.T) {
	attempts := 0

	err := withRetry(context.Background(), 4, []time.Duration{0, 0, 0, 0}, func() error {
		attempts++
		if attempts < 3 {
			return &testError{} // retriable error
		}
		return nil // success on 3rd try
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "Should have retried twice before success")
}

func TestWithRetry_FailsAfterMaxAttempts(t *testing.T) {
	attempts := 0

	err := withRetry(context.Background(), 4, []time.Duration{0, 0, 0, 0}, func() error {
		attempts++
		return &testError{} // always retriable error
	})

	assert.Error(t, err)
	assert.Equal(t, 4, attempts, "Should retry max attempts")
}

func TestRetryMiddleware_FailMaxAttempts(t *testing.T) {
	attempts := 0

	// This handler simulates a retriable error by triggering the error on the BufferedResponseWriter
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// We cannot cast w to BufferedResponseWriter here, but since
		// your middleware creates it internally, we can't set error here.
		// So instead, simulate the error by wrapping the middleware:

		// Just write a status code for now.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fail"))
	})

	// We override RetryMiddleware temporarily with our own that injects error in BufferedResponseWriter
	retryWithErrorInjection := func(next http.Handler) http.Handler {
		const maxAttempts = 4
		delays := []time.Duration{0, 0, 0, 0}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var brw *BufferedResponseWriter

			err := withRetry(r.Context(), maxAttempts, delays, func() error {
				brw = NewBufferedResponseWriter()
				next.ServeHTTP(brw, r)
				// Simulate retriable error on all attempts
				return &testError{}
			})

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			brw.flushTo(w)
		})
	}

	middleware := retryWithErrorInjection(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <----- Close response body here

	assert.Equal(t, 4, attempts, "Should retry maximum attempts (4)")
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
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
			name: "pgconn retriable error with code 08xxx",
			err: &pgconn.PgError{
				Code: "08003", // connection does not exist (starts with 08)
			},
			want: true,
		},
		{
			name: "pgconn non-retriable error",
			err: &pgconn.PgError{
				Code: "23505", // unique violation
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
			name: "os.PathError with other error",
			err: &os.PathError{
				Err: errors.New("some other error"),
			},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		delays      []time.Duration
		fn          func() error
		wantErr     bool
		cancelCtx   bool // whether to cancel context mid-way
		wantCalls   int  // expected calls to fn
	}{
		{
			name:        "succeeds first try",
			maxAttempts: 3,
			delays:      []time.Duration{0, 0},
			fn: func() error {
				return nil
			},
			wantErr:   false,
			wantCalls: 1,
		},
		{
			name:        "fails twice, succeeds third try",
			maxAttempts: 4,
			delays:      []time.Duration{0, 0, 0},
			fn: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts < 3 {
						return errors.New("fail")
					}
					return nil
				}
			}(),
			wantErr:   false,
			wantCalls: 3,
		},
		{
			name:        "fails all attempts",
			maxAttempts: 3,
			delays:      []time.Duration{0, 0},
			fn: func() error {
				return errors.New("fail")
			},
			wantErr:   true,
			wantCalls: 3,
		},
		{
			name:        "context canceled before completion",
			maxAttempts: 5,
			delays:      []time.Duration{time.Second, time.Second, time.Second, time.Second},
			cancelCtx:   true,
			fn: func() error {
				return errors.New("fail")
			},
			wantErr:   true,
			wantCalls: 1, // only one call before ctx canceled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.cancelCtx {
				ctx, cancel = context.WithCancel(ctx)
				// cancel context after short delay to interrupt retries
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			callCount := 0
			err := withRetry(ctx, tt.maxAttempts, tt.delays, func() error {
				callCount++
				return tt.fn()
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalls, callCount)
		})
	}
}

func TestRetryMiddleware_RetriableError_Response503(t *testing.T) {
	attempts := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		brw, ok := w.(*BufferedResponseWriter)
		if !ok {
			t.Fatal("expected *BufferedResponseWriter")
		}

		// Always set a retriable error to force retry attempts
		brw.SetError(&pgconn.PgError{Code: "08006"})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("error response"))
	})

	middleware := RetryMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close() // <----- Close response body here
	body := rec.Body.String()

	assert.Equal(t, 4, attempts, "should retry max attempts")
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "should return 503 after retries fail")
	assert.Equal(t, "", body, "should not write body on failure") // body is empty because we return early with 503
}
