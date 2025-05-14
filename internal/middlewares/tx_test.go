package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetTx_WithTxInContext(t *testing.T) {
	tx := &sqlx.Tx{}
	ctx := setTx(context.Background(), tx)
	result := GetTx(ctx)
	assert.Equal(t, tx, result, "Expected to get the same transaction from context")
}

func TestGetTx_NoTxInContext(t *testing.T) {
	ctx := context.Background()
	result := GetTx(ctx)
	assert.Nil(t, result, "Expected nil when no transaction is in context")
}

func TestTxMiddleware_NoDB(t *testing.T) {
	handler := TxMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTxMiddleware_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' occurred when opening mock database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	handler := TxMiddleware(sqlxDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTx(r.Context())
		assert.NotNil(t, tx, "Transaction should be set in the context")
		if err := tx.Commit(); err != nil {
			logger.Log.Error("TxMiddleware: commit failed", zap.Error(err))
			_ = tx.Rollback()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}))
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTxMiddleware_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' occurred when opening mock database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()
	handler := TxMiddleware(sqlxDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTx(r.Context())
		assert.NotNil(t, tx, "Transaction should be set in the context")
		w.WriteHeader(http.StatusOK)
	}))
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTxMiddleware_BeginTransactionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' occurred when opening mock database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin().WillReturnError(fmt.Errorf("failed to begin transaction"))
	handler := TxMiddleware(sqlxDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTxMiddleware_PanicHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' occurred when opening mock database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()
	handler := TxMiddleware(sqlxDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTx(r.Context())
		assert.NotNil(t, tx, "Transaction should be set in the context")
		panic("Test panic")
	}))
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "Test panic", r)
		}
	}()
	handler.ServeHTTP(rr, req)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
