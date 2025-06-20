package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxMiddleware_NoDB(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := TxMiddleware(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.True(t, handlerCalled, "handler should be called when DB is nil")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTxMiddleware_BeginFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	mw := TxMiddleware(sqlxDB)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should NOT be called if begin fails")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_CommitSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectCommit()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		tx := GetTx(r.Context())
		assert.NotNil(t, tx)
		w.WriteHeader(http.StatusOK)
	})

	mw := TxMiddleware(sqlxDB)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_RollbackOnStatusCode400OrMore(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectRollback()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	mw := TxMiddleware(sqlxDB)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_CommitError_SetsStatusCodeIfNoHeader(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	// Case 1: handler does NOT write header -> middleware sets 500
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no header write here
	})

	mw := TxMiddleware(sqlxDB)
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr1 := httptest.NewRecorder()

	mw(handler1).ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusInternalServerError, rr1.Code)

	require.NoError(t, mock.ExpectationsWereMet())

	// Reset mock for second case
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	// Case 2: handler writes header already -> middleware does NOT override status code
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()

	mw(handler2).ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code) // middleware does not override

	require.NoError(t, mock.ExpectationsWereMet())
}
