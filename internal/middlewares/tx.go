package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/tx"
)

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
