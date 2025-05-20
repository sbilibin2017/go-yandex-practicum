package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// определим безопасный тип ключа для context
type txKeyType struct{}

var txKey = txKeyType{}

func TestTxMiddleware_NilTxSetter_CallsNext(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mw := TxMiddleware(sqlxDB, nil)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestTxMiddleware_NilDB_CallsNext(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mw := TxMiddleware(nil, func(ctx context.Context, tx *sqlx.Tx) context.Context {
		return ctx
	})

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

	txSetter := func(ctx context.Context, tx *sqlx.Tx) context.Context {
		return context.WithValue(ctx, txKey, tx)
	}

	middleware := TxMiddleware(sqlxDB, txSetter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := r.Context().Value(txKey)
		assert.NotNil(t, tx)
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

	txSetter := func(ctx context.Context, tx *sqlx.Tx) context.Context {
		return ctx
	}

	middleware := TxMiddleware(sqlxDB, txSetter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on begin error")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
