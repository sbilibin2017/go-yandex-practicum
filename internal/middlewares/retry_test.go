package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- BufferedResponseWriter tests ---

func TestBufferedResponseWriter_WriteAndHeader(t *testing.T) {
	brw := NewBufferedResponseWriter()

	// Initially header is empty, status 200 OK
	assert.Equal(t, http.StatusOK, brw.statusCode)
	assert.False(t, brw.wroteHeader)
	assert.NotNil(t, brw.Header())

	// WriteHeader sets status code and wroteHeader flag
	brw.WriteHeader(http.StatusAccepted)
	assert.Equal(t, http.StatusAccepted, brw.statusCode)
	assert.True(t, brw.wroteHeader)

	// WriteHeader called again does nothing
	brw.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusAccepted, brw.statusCode)

	// Write writes to buffer and triggers WriteHeader if not called
	brw2 := NewBufferedResponseWriter()
	n, err := brw2.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.True(t, brw2.wroteHeader)
	assert.Equal(t, http.StatusOK, brw2.statusCode)
	assert.Equal(t, "hello", brw2.body.String())
}

func TestBufferedResponseWriter_SetError(t *testing.T) {
	brw := NewBufferedResponseWriter()
	err := errors.New("test error")
	brw.SetError(err)
	assert.Equal(t, err, brw.err)
}

func TestBufferedResponseWriter_FlushTo(t *testing.T) {
	brw := NewBufferedResponseWriter()
	brw.Header().Set("X-Foo", "bar")
	brw.WriteHeader(http.StatusCreated)
	brw.Write([]byte("response body"))

	rr := httptest.NewRecorder()
	brw.flushTo(rr)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "bar", rr.Header().Get("X-Foo"))
	assert.Equal(t, "response body", rr.Body.String())
}

// --- isRetriableError tests ---

func TestIsRetriableError(t *testing.T) {
	// nil error
	assert.False(t, isRetriableError(nil))

	// PgError code 08xxx -> retriable
	pgErr := &pgconn.PgError{Code: "08003"}
	assert.True(t, isRetriableError(pgErr))

	// PgError code not starting with 08 -> not retriable
	pgErr.Code = "23505"
	assert.False(t, isRetriableError(pgErr))

	// os.PathError with EAGAIN -> retriable
	pathErr := &os.PathError{Err: syscall.EAGAIN}
	assert.True(t, isRetriableError(pathErr))

	// os.PathError with EWOULDBLOCK -> retriable
	pathErr.Err = syscall.EWOULDBLOCK
	assert.True(t, isRetriableError(pathErr))

	// os.PathError with other error -> not retriable
	pathErr.Err = syscall.ECONNRESET
	assert.False(t, isRetriableError(pathErr))

	// some other error -> not retriable
	assert.False(t, isRetriableError(errors.New("random error")))
}

// --- RetryMiddleware tests ---

// testHandler uses handlerFn for custom ServeHTTP logic
type testHandler struct {
	callCount int
	errToSet  error
	handlerFn func(w http.ResponseWriter, r *http.Request)
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callCount++
	if h.handlerFn != nil {
		h.handlerFn(w, r)
		return
	}

	if brw, ok := w.(*BufferedResponseWriter); ok {
		brw.SetError(h.errToSet)
	}
}

func TestRetryMiddleware_NoError_NoRetry(t *testing.T) {
	h := &testHandler{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := RetryMiddleware(h)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 1, h.callCount)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRetryMiddleware_RetryThenSuccess(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "08003"} // retriable
	h := &testHandler{}

	call := 0
	h.handlerFn = func(w http.ResponseWriter, r *http.Request) {
		call++
		if brw, ok := w.(*BufferedResponseWriter); ok {
			if call < 3 {
				brw.SetError(pgErr)
			}
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := RetryMiddleware(h)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 3, call)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRetryMiddleware_RetryExhausted(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "08003"} // retriable
	h := &testHandler{
		errToSet: pgErr,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := RetryMiddleware(h)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 4, h.callCount)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestRetryMiddleware_NonRetriableError(t *testing.T) {
	h := &testHandler{
		errToSet: errors.New("non retriable"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := RetryMiddleware(h)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 1, h.callCount)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRetryMiddleware_ContextCanceled(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "08003"} // retriable
	h := &testHandler{
		errToSet: pgErr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler := RetryMiddleware(h)
	handler.ServeHTTP(rr, req)

	assert.Less(t, h.callCount, 5)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
