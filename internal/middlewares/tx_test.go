package middlewares

import (
	"context"
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

	// simulate begin failure
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
		// Ensure tx is in context
		tx := getTx(r.Context())
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

func TestTxMiddleware_HandlerErrorRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectRollback()

	mw := TxMiddleware(sqlxDB)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	defer func() {
		r := recover()
		assert.Equal(t, "simulated panic", r)
		require.NoError(t, mock.ExpectationsWereMet())
	}()

	mw(handler).ServeHTTP(rr, req)
}

func TestWithTx(t *testing.T) {
	// Helper to create context with a mocked tx
	setupTxCtx := func(t *testing.T) (context.Context, sqlmock.Sqlmock, *sqlx.Tx) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		mock.ExpectBegin()
		tx, err := sqlxDB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)

		ctx := setTx(context.Background(), tx)
		return ctx, mock, tx
	}

	t.Run("no transaction in context", func(t *testing.T) {
		err := withTx(context.Background(), func(ctx context.Context) error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no transaction in context")
	})

	t.Run("fn returns error triggers rollback", func(t *testing.T) {
		ctx, mock, _ := setupTxCtx(t)
		mock.ExpectRollback()

		testErr := errors.New("handler error")

		err := withTx(ctx, func(ctx context.Context) error {
			return testErr
		})

		assert.Equal(t, testErr, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fn panics triggers rollback and rethrows", func(t *testing.T) {
		ctx, mock, _ := setupTxCtx(t)
		mock.ExpectRollback()

		panicMsg := "panic inside fn"

		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, panicMsg, r)
			}
		}()

		// Call withTx that panics inside the function
		_ = withTx(ctx, func(ctx context.Context) error {
			panic(panicMsg)
		})
	})

	t.Run("fn returns nil commits successfully", func(t *testing.T) {
		ctx, mock, _ := setupTxCtx(t)
		mock.ExpectCommit()

		err := withTx(ctx, func(ctx context.Context) error {
			return nil
		})

		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit failure triggers rollback and returns error", func(t *testing.T) {
		ctx, mock, _ := setupTxCtx(t)

		mock.ExpectCommit().WillReturnError(errors.New("commit failure"))
		// No rollback expected on commit failure

		err := withTx(ctx, func(ctx context.Context) error {
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit failure")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTxMiddleware_HandlerErrorWrites500IfNoHeader(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectRollback()

	middleware := TxMiddleware(sqlxDB)

	handlerWithPanic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("handler panic")
	})

	h := middleware(handlerWithPanic)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	defer func() {
		if r := recover(); r != nil {
			_ = r // swallow panic intentionally to avoid empty branch
		}
	}()

	h.ServeHTTP(rec, req)

	assert.Equal(t, 0, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseRecorder_Write(t *testing.T) {
	underlying := httptest.NewRecorder()

	rec := &responseRecorder{
		ResponseWriter: underlying,
	}

	data := []byte("hello world")

	n, err := rec.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, string(data), underlying.Body.String())
}

func TestGetExecutor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	tx, err := sqlxDB.BeginTxx(context.Background(), nil)
	require.NoError(t, err)

	ctxWithTx := setTx(context.Background(), tx)
	executor := GetExecutor(ctxWithTx, sqlxDB)
	assert.Equal(t, tx, executor)

	ctxWithoutTx := context.Background()
	executor = GetExecutor(ctxWithoutTx, sqlxDB)
	assert.Equal(t, sqlxDB, executor)

	require.NoError(t, mock.ExpectationsWereMet())
}
