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

func dummyTxSetter(ctx context.Context, tx *sqlx.Tx) context.Context {
	return ctx
}

func TestTxMiddleware_DBNil(t *testing.T) {
	handlerCalled := false
	config := TxMiddleware(nil, dummyTxSetter)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	config(nextHandler).ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTxMiddleware_BeginTxFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin().WillReturnError(assert.AnError)

	middleware := TxMiddleware(sqlxDB, dummyTxSetter)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called if BeginTxx fails")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectCommit()

	called := false
	middleware := TxMiddleware(sqlxDB, func(ctx context.Context, tx *sqlx.Tx) context.Context {
		called = true
		return ctx
	})

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotNil(t, r.Context())
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_HandlerPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectRollback()

	middleware := TxMiddleware(sqlxDB, dummyTxSetter)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		middleware(nextHandler).ServeHTTP(w, req)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

type txTestKey string

var k txTestKey = "test"

func TestWithTx_DBIsNil(t *testing.T) {
	var called bool
	err := withTx(
		context.Background(),
		nil,
		func(ctx context.Context, tx *sqlx.Tx) context.Context {
			t.Fatal("txSetter should not be called when db is nil")
			return ctx
		},
		func(ctx context.Context, tx *sqlx.Tx) error {
			called = true
			return nil
		},
	)
	require.NoError(t, err)
	assert.False(t, called, "fn should not be called when db is nil")
}

func TestWithTx_BeginFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
	err = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
		t.Fatal("should not be called")
		return nil
	})
	assert.EqualError(t, err, "begin failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()
	err = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_FuncReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()
	testErr := errors.New("some error")
	err = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
		return testErr
	})
	assert.ErrorIs(t, err, testErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_CommitFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
	err = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
		return nil
	})
	assert.EqualError(t, err, "commit failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))
	mainErr := errors.New("main error")
	err = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
		return mainErr
	})
	assert.ErrorContains(t, err, "main error")
	assert.ErrorContains(t, err, "rollback failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_Panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()
	assert.Panics(t, func() {
		_ = withTx(context.Background(), sqlxDB, dummyTxSetter, func(ctx context.Context, tx *sqlx.Tx) error {
			panic("unexpected panic")
		})
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_TxSetterInjectsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()

	var extractedTx *sqlx.Tx

	err = withTx(context.Background(), sqlxDB,
		func(ctx context.Context, tx *sqlx.Tx) context.Context {
			return context.WithValue(ctx, k, tx)
		},
		func(ctx context.Context, tx *sqlx.Tx) error {
			v := ctx.Value(k)
			require.NotNil(t, v, "tx should be injected into context")
			var ok bool
			extractedTx, ok = v.(*sqlx.Tx)
			require.True(t, ok, "value in context should be *sqlx.Tx")
			require.Equal(t, tx, extractedTx, "injected tx should match argument tx")
			return nil
		},
	)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetTxAndGetTx_WithSqlMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	tx, err := sqlxDB.Beginx()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	ctx := SetTx(context.Background(), tx)
	txFromCtx := GetTx(ctx)
	assert.Equal(t, tx, txFromCtx, "транзакция из контекста должна совпадать с исходной")
}

func TestGetTx_EmptyContext_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	tx := GetTx(ctx)
	assert.Nil(t, tx, "для пустого контекста должно вернуться nil")
}

func TestGetTx_WrongType_ReturnsNil(t *testing.T) {
	type anotherKeyType struct{}
	key := anotherKeyType{}
	ctx := context.WithValue(context.Background(), key, "not a tx")
	tx := GetTx(ctx)
	assert.Nil(t, tx, "если ключ отличается, значение не должно быть извлечено")
}
