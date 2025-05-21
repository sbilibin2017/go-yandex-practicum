package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestTxMiddleware_NilDB_CallsNext(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mw := TxMiddleware(nil)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestTxMiddleware_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectCommit()

	middleware := TxMiddleware(sqlxDB)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := r.Context().Value(txContextKey{})
		assert.NotNil(t, tx)
		_, ok := tx.(*sqlx.Tx)
		assert.True(t, ok)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_BeginFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin().WillReturnError(assert.AnError)

	middleware := TxMiddleware(sqlxDB)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on begin error")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
