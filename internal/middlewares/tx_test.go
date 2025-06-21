package middlewares_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
)

// define custom type for context keys
type ctxKey string

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	cleanup := func() {
		sqlxDB.Close()
	}

	return sqlxDB, mock, cleanup
}

func TestTxMiddleware_SuccessfulCommit(t *testing.T) {
	sqlxDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit()

	middleware, err := middlewares.TxMiddleware(middlewares.WithDB(sqlxDB))
	require.NoError(t, err)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_BeginTxFails(t *testing.T) {
	sqlxDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectBegin().WillReturnError(errors.New("begin tx failed"))

	middleware, err := middlewares.TxMiddleware(middlewares.WithDB(sqlxDB))
	require.NoError(t, err)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called if begin tx fails")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_CommitFails_TriggersRollback(t *testing.T) {
	sqlxDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
	mock.ExpectRollback()

	middleware, err := middlewares.TxMiddleware(middlewares.WithDB(sqlxDB))
	require.NoError(t, err)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTxMiddleware_NoDBConfigured_CallsNextDirectly(t *testing.T) {
	middleware, err := middlewares.TxMiddleware() // no DB option
	require.NoError(t, err)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no db"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no db", rec.Body.String())
}

func TestTxMiddleware_TxSetterInjectsTx(t *testing.T) {
	sqlxDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit()

	var capturedTx *sqlx.Tx

	txSetter := func(ctx context.Context, tx *sqlx.Tx) context.Context {
		capturedTx = tx
		return context.WithValue(ctx, ctxKey("tx"), tx)
	}

	middleware, err := middlewares.TxMiddleware(
		middlewares.WithDB(sqlxDB),
		middlewares.WithTxSetter(txSetter),
	)
	require.NoError(t, err)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txFromCtx := r.Context().Value(ctxKey("tx"))
		assert.NotNil(t, txFromCtx)
		assert.Equal(t, capturedTx, txFromCtx)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewTxMiddlewareConfig_NoOptions(t *testing.T) {
	cfg := middlewares.NewTxMiddlewareConfig()
	assert.Nil(t, cfg.DB)
	assert.Nil(t, cfg.TxOptions)
	assert.Nil(t, cfg.TxSetter)
}

func TestWithDB(t *testing.T) {
	db := &sqlx.DB{} // dummy instance
	cfg := middlewares.NewTxMiddlewareConfig(middlewares.WithDB(db))
	assert.Equal(t, db, cfg.DB)
}

func TestWithTxOptions(t *testing.T) {
	opts := &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: true}
	cfg := middlewares.NewTxMiddlewareConfig(middlewares.WithTxOptions(opts))
	assert.Equal(t, opts, cfg.TxOptions)
}

func TestWithTxSetter(t *testing.T) {
	setter := func(ctx context.Context, tx *sqlx.Tx) context.Context {
		return context.WithValue(ctx, ctxKey("key"), "value")
	}
	cfg := middlewares.NewTxMiddlewareConfig(middlewares.WithTxSetter(setter))
	assert.NotNil(t, cfg.TxSetter)

	ctx := context.Background()
	newCtx := cfg.TxSetter(ctx, nil)
	val := newCtx.Value(ctxKey("key"))
	assert.Equal(t, "value", val)
}

func TestNewTxMiddlewareConfig_MultipleOptions(t *testing.T) {
	db := &sqlx.DB{}
	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	setter := func(ctx context.Context, tx *sqlx.Tx) context.Context { return ctx }

	cfg := middlewares.NewTxMiddlewareConfig(
		middlewares.WithDB(db),
		middlewares.WithTxOptions(opts),
		middlewares.WithTxSetter(setter),
	)

	assert.NotNil(t, cfg)
}
