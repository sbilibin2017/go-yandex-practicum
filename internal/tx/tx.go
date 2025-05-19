package tx

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// WithTx выполняет функцию fn в рамках транзакции базы данных.
// ctx — исходный контекст.
// db — указатель на sqlx.DB, из которого создаётся транзакция.
// fn — функция, принимающая контекст и выполняющая действия в транзакции.
// txSetter — функция для записи транзакции в контекст (например, SetTxToContext).
//
// Функция начинает новую транзакцию с помощью db.BeginTxx,
// затем вызывает fn с обновлённым контекстом, содержащим транзакцию.
// Если fn возвращает ошибку, транзакция откатывается.
// Если fn завершается успешно, транзакция коммитится.
//
// Возвращает ошибку, возникшую при создании транзакции, выполнении fn,
// откате или коммите.
func WithTx(
	ctx context.Context,
	db *sqlx.DB,
	fn func(ctx context.Context) error,
	txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context,
) error {
	if db == nil {
		return nil
	}

	if txSetter == nil {
		return nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	ctx = txSetter(ctx, tx)

	err = fn(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
