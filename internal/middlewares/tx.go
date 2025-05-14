package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

func TxMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := db.BeginTxx(r.Context(), nil)
			if err != nil {
				logger.Log.Error("TxMiddleware: failed to begin transaction", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			defer func() {
				if rec := recover(); rec != nil {
					logger.Log.Error("TxMiddleware: panic occurred, rolling back", zap.Any("recover", rec))
					_ = tx.Rollback()
					panic(rec)
				}
			}()

			ctx := setTx(r.Context(), tx)
			reqWithTx := r.WithContext(ctx)

			next.ServeHTTP(w, reqWithTx)

			if err := tx.Commit(); err != nil {
				logger.Log.Error("TxMiddleware: commit failed", zap.Error(err))
				_ = tx.Rollback()
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})
	}
}

type txContextKey struct{}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

func setTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}
