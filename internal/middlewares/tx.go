package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/tx"
)

// TxMiddleware — HTTP middleware, оборачивающий обработчики в SQL-транзакцию.
//
// Транзакция создаётся перед вызовом обработчика и коммитится по завершению,
// либо откатывается при возникновении ошибки.
//
// Использует переданную функцию withTxFunc для управления жизненным циклом транзакции.
//
// Параметры:
//   - db: *sqlx.DB — пул соединений к базе данных
//   - withTxFunc: функция, запускающая выполнение в транзакции (см. internal/tx.WithTx)
//
// Пример использования:
//
//	http.Handle("/", TxMiddleware(db, tx.WithTx)(handler))
func TxMiddleware(
	db *sqlx.DB,
	withTxFunc func(
		ctx context.Context,
		db *sqlx.DB,
		fn func(ctx context.Context) error,
		txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context,
	) error,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			err := withTxFunc(ctx, db, func(ctx context.Context) error {
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return nil
			}, tx.SetTxToContext)

			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})
	}
}
