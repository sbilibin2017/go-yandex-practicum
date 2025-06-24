package contexts

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGetTxFromContext(t *testing.T) {
	// Create a dummy *sqlx.Tx pointer (can be nil or a mocked tx)
	dummyTx := &sqlx.Tx{}

	ctx := context.Background()

	// Set the tx in context
	ctxWithTx := SetTxToContext(ctx, dummyTx)

	// Retrieve the tx from context
	tx, ok := GetTxFromContext(ctxWithTx)

	assert.True(t, ok, "expected to find tx in context")
	assert.Equal(t, dummyTx, tx, "tx retrieved should be the same as the one set")

	// Test getting from a context with no tx set
	_, ok = GetTxFromContext(ctx)
	assert.False(t, ok, "expected no tx in empty context")
}
