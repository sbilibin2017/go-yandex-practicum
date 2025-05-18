package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBRetryMiddleware_Success(t *testing.T) {
	attempts := []time.Duration{time.Millisecond * 10}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	middleware := DBRetryMiddleware(
		withRetryMock,
		attempts,
		func(err error) bool { return false },
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", strings.TrimSpace(w.Body.String()))
}

func TestDBRetryMiddleware_WithRetrySuccess(t *testing.T) {
	attempts := []time.Duration{time.Millisecond * 10, time.Millisecond * 10}
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			panic("temporary error")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("recovered"))
	})

	middleware := DBRetryMiddleware(
		withRetryMock,
		attempts,
		func(err error) bool {
			return err.Error() == "handler panic occurred"
		},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "recovered", strings.TrimSpace(w.Body.String()))
	assert.Equal(t, 2, callCount)
}

func TestDBRetryMiddleware_NonRetriableError(t *testing.T) {
	attempts := []time.Duration{time.Millisecond * 10}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
		panic("fatal error")
	})

	middleware := DBRetryMiddleware(
		withRetryMock,
		attempts,
		func(err error) bool {
			return false
		},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "bad request", strings.TrimSpace(w.Body.String()))
}

func TestDBRetryMiddleware_FailsAfterRetries(t *testing.T) {
	attempts := []time.Duration{
		time.Millisecond * 10,
		time.Millisecond * 10,
		time.Millisecond * 10,
	}
	callCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		http.Error(w, "fail", http.StatusInternalServerError)
		panic("handler panic occurred")
	})

	middleware := DBRetryMiddleware(
		withRetryMock,
		attempts,
		func(err error) bool {
			return err.Error() == "handler panic occurred"
		},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "fail", strings.TrimSpace(w.Body.String()))
	assert.Equal(t, len(attempts)+1, callCount)
}

func withRetryMock(
	ctx context.Context,
	attempts []time.Duration,
	fn func(ctx context.Context) error,
	isRetriableErrorFuncs ...func(err error) bool,
) error {
	var lastErr error

	for i := 0; i <= len(attempts); i++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}

		isRetriable := false
		for _, f := range isRetriableErrorFuncs {
			if f(err) {
				isRetriable = true
				break
			}
		}

		if !isRetriable {
			return err
		}

		lastErr = err

		if i < len(attempts) {
			time.Sleep(attempts[i])
		}
	}

	return lastErr
}
