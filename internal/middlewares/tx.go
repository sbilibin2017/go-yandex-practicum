package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func TxMiddleware(db *sqlx.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			tx, err := db.BeginTxx(ctx, nil)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			ctx = SetTx(ctx, tx)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)

			if err := tx.Commit(); err != nil {
				_ = tx.Rollback()
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		})
	}
}

type txContextKey struct{}

func SetTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}
