package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
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
