package contexts

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txContextKey struct{}

func SetTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}
