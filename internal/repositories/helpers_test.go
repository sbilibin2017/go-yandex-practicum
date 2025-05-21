package repositories

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestGetExecutor_ReturnsTxIfExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	tx, err := sqlxDB.Beginx()
	assert.NoError(t, err)

	ctx := middlewares.SetTx(context.Background(), tx)

	executor := getExecutor(ctx, sqlxDB)

	assert.Equal(t, tx, executor)

	_ = tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExecutor_ReturnsDBIfTxIsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	ctx := context.Background()

	executor := getExecutor(ctx, sqlxDB)

	assert.Equal(t, sqlxDB, executor)
	assert.NoError(t, mock.ExpectationsWereMet())
}
