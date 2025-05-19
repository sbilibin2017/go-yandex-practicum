package tx

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// txContextKey — тип для ключа контекста, используемого для хранения транзакции.
// Используется для предотвращения коллизий ключей в контексте.
type txContextKey struct{}

// SetTxToContext возвращает новый контекст, в который помещена транзакция tx.
// ctx — исходный контекст.
// tx — транзакция *sqlx.Tx, которую необходимо сохранить в контексте.
// Позволяет передавать транзакцию между функциями через контекст.
func SetTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// GetTxFromContext извлекает транзакцию *sqlx.Tx из контекста ctx.
// Возвращает транзакцию, если она была установлена, иначе nil.
func GetTxFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}
