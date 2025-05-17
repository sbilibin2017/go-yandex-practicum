package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBRetryMiddlewareStandard_Success(t *testing.T) {
	calls := 0

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK from inner handler"))
	})

	handlerWithMiddleware := DBRetryMiddleware(innerHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handlerWithMiddleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK from inner handler", rec.Body.String())
	assert.Equal(t, 1, calls)
}

func TestDBRetryMiddleware_SuccessFirstTry(t *testing.T) {
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) error {
		calls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	dbRetryMiddleware(handler)(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, calls)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestDBRetryMiddleware_RetryAndFail(t *testing.T) {
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) error {
		calls++
		return newRetriablePGError("08006") // Connection failure (retriable)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	dbRetryMiddleware(handler)(rec, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.GreaterOrEqual(t, calls, 4) // initial + 3 retries
	// Проверяем, что время примерно больше или равно сумме задержек
	assert.GreaterOrEqual(t, duration, 9*time.Second)
}

func TestDBRetryMiddleware_RetryAndSucceed(t *testing.T) {
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) error {
		calls++
		if calls < 3 {
			return newRetriablePGError("08001") // Connection exception, retriable
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Recovered"))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	dbRetryMiddleware(handler)(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Recovered", rec.Body.String())
	assert.Equal(t, 3, calls)
}

func TestDBRetryMiddleware_NonRetriableError(t *testing.T) {
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) error {
		calls++
		return errors.New("fatal DB error") // Non-retriable
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	dbRetryMiddleware(handler)(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 1, calls)
}

func newRetriablePGError(code string) error {
	return &pgconn.PgError{
		Code: code,
	}
}

func makePgError(code string) error {
	return &pgconn.PgError{Code: code}
}

func TestWithRetry_SuccessFirstTry(t *testing.T) {
	called := 0
	err := withRetry(context.Background(), []time.Duration{time.Millisecond}, func(ctx context.Context) error {
		called++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestWithRetry_SuccessAfterRetry(t *testing.T) {
	called := 0
	err := withRetry(context.Background(), []time.Duration{time.Millisecond, time.Millisecond}, func(ctx context.Context) error {
		called++
		if called < 2 {
			return makePgError("08000") // retriable error code
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithRetry_NonRetriableError(t *testing.T) {
	called := 0
	expectedErr := errors.New("non-retriable error")
	err := withRetry(context.Background(), []time.Duration{time.Millisecond}, func(ctx context.Context) error {
		called++
		return expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 1, called)
}

func TestWithRetry_ExceedsRetries(t *testing.T) {
	called := 0
	retriableErr := makePgError("08000")
	delays := []time.Duration{time.Millisecond, time.Millisecond}

	err := withRetry(context.Background(), delays, func(ctx context.Context) error {
		called++
		return retriableErr
	})
	assert.ErrorIs(t, err, retriableErr)
	assert.Equal(t, len(delays)+1, called)
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := 0
	delays := []time.Duration{100 * time.Millisecond, 100 * time.Millisecond}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := withRetry(ctx, delays, func(ctx context.Context) error {
		called++
		return makePgError("08000") // retriable error
	})

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Less(t, called, len(delays)+1)
}

func TestIsRetriableDBError(t *testing.T) {
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
			name: "non-pg error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "pg error with code starting 08",
			err:  &pgconn.PgError{Code: "08003"},
			want: true,
		},
		{
			name: "pg error with code not starting 08",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "pg error with short code",
			err:  &pgconn.PgError{Code: "0"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableDBError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
