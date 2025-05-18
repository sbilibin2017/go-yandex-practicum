package tx

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestWithTx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectCommit()
	ctx := context.Background()
	called := false
	err = WithTx(ctx, sqlxDB, func(ctx context.Context) error {
		called = true
		txObj := GetTxFromContext(ctx)
		assert.NotNil(t, txObj)
		return nil
	}, SetTxToContext)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_FnReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectRollback()
	ctx := context.Background()
	testErr := errors.New("some error")
	err = WithTx(ctx, sqlxDB, func(ctx context.Context) error {
		return testErr
	}, SetTxToContext)
	assert.ErrorIs(t, err, testErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_NilDB(t *testing.T) {
	ctx := context.Background()
	err := WithTx(ctx, nil, func(ctx context.Context) error {
		return nil
	}, SetTxToContext)
	assert.NoError(t, err)
}

func TestWithTx_NilTxSetter(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	ctx := context.Background()
	err = WithTx(ctx, sqlxDB, func(ctx context.Context) error {
		return nil
	}, nil)
	assert.NoError(t, err)
}

func TestWithTx_BeginTxxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))
	ctx := context.Background()
	err = WithTx(ctx, sqlxDB, func(ctx context.Context) error {
		t.Fatal("fn should not be called")
		return nil
	}, SetTxToContext)
	assert.Error(t, err)
	assert.EqualError(t, err, "begin error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
