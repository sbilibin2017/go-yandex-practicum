package repositories

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetExecutor_ReturnsTxIfExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Настраиваем ожидание вызова Begin
	mock.ExpectBegin()

	// Начинаем транзакцию через sqlx
	sqlxTx, err := sqlxDB.Beginx()
	assert.NoError(t, err)

	// txGetter возвращает транзакцию
	txGetter := func(ctx context.Context) *sqlx.Tx {
		return sqlxTx
	}

	ctx := context.Background()

	executor := getExecutor(ctx, sqlxDB, txGetter)

	assert.Equal(t, sqlxTx, executor)

	// откатываем транзакцию для очистки
	_ = sqlxTx.Rollback()

	// Проверяем, что все ожидания были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExecutor_ReturnsDBIfTxNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// txGetter возвращает nil
	txGetter := func(ctx context.Context) *sqlx.Tx {
		return nil
	}

	ctx := context.Background()

	executor := getExecutor(ctx, sqlxDB, txGetter)

	assert.Equal(t, sqlxDB, executor)

	assert.NoError(t, mock.ExpectationsWereMet())
}
