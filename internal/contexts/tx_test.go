package contexts

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSetTxToContextAndGetTxFromContext(t *testing.T) {
	// Создаем мок базы и sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Оборачиваем в sqlx.DB
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Начинаем транзакцию
	mock.ExpectBegin()
	txObj, err := sqlxDB.Beginx()
	assert.NoError(t, err)

	// Кладём транзакцию в контекст
	ctx := context.Background()
	ctxWithTx := SetTx(ctx, txObj)

	// Получаем транзакцию из контекста
	txFromCtx := GetTx(ctxWithTx)

	assert.NotNil(t, txFromCtx)
	assert.Equal(t, txObj, txFromCtx)

	// Проверяем, что из пустого контекста вернется nil
	txFromEmptyCtx := GetTx(context.Background())
	assert.Nil(t, txFromEmptyCtx)

	// Завершаем транзакцию
	mock.ExpectCommit()
	err = txObj.Commit()
	assert.NoError(t, err)

	// Проверяем, что все ожидания sqlmock выполнены
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
