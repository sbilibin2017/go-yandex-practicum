package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/tx"
)

func TestTxMiddleware_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		txFromCtx := tx.GetTxFromContext(r.Context())
		assert.NotNil(t, txFromCtx)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	middleware := middlewares.TxMiddleware(sqlxDB, tx.WithTx)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, handlerCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_WithTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()
	middleware := middlewares.TxMiddleware(sqlxDB, tx.WithTx)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	errReturned := tx.WithTx(req.Context(), sqlxDB, func(ctx context.Context) error {
		return errors.New("handler error")
	}, tx.SetTxToContext)
	assert.Error(t, errReturned)
	assert.Equal(t, "handler error", errReturned.Error())
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})).ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))
	middleware := middlewares.TxMiddleware(sqlxDB, tx.WithTx)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})
	middleware(handler).ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.False(t, handlerCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}
