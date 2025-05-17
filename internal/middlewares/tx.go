package middlewares

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

func TxMiddleware(
	db *sqlx.DB,
	txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			err := withTx(r.Context(), db, txSetter, func(ctx context.Context, tx *sqlx.Tx) error {
				reqWithTx := r.WithContext(ctx)
				next.ServeHTTP(w, reqWithTx)
				return nil
			})

			if err != nil {
				logger.Log.Error("TxMiddleware: transaction failed", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		})
	}
}

func withTx(
	ctx context.Context,
	db *sqlx.DB,
	txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context,
	fn func(ctx context.Context, tx *sqlx.Tx) error,
) error {
	if db == nil {
		return nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if txSetter != nil {
		ctx = txSetter(ctx, tx)
	}

	err = fn(ctx, tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

type txContextKey struct{}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

func SetTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}
