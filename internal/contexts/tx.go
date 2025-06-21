package contexts

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// txContextKey is an unexported type used as the key for storing the sqlx.Tx
// object in a context.Context to avoid key collisions.
type txContextKey struct{}

// SetTxToContext returns a new context with the given sqlx.Tx transaction
// stored in it. This can be used to pass a database transaction through
// various layers of the application.
func SetTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// GetTxFromContext retrieves a sqlx.Tx transaction from the provided context.
// It returns the transaction and a boolean indicating whether the transaction
// was present in the context and had the correct type.
func GetTxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx, ok
}
