package tx

import (
	"context"

	"github.com/jmoiron/sqlx"
)

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
