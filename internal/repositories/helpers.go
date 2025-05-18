package repositories

import (
	"context"
	"os"

	"github.com/jmoiron/sqlx"
)

func withFileSync(
	file *os.File,
	fn func(*os.File) error,
) error {
	if err := file.Sync(); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if err := fn(file); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

type namedPreparer interface {
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

func getExecutor(
	ctx context.Context,
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) namedPreparer {
	if tx := txGetter(ctx); tx != nil {
		return tx
	}
	return db
}
