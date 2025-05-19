package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllDBRepository struct {
	db         *sqlx.DB
	txProvider func(ctx context.Context) *sqlx.Tx
}

func NewMetricListAllDBRepository(
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
) *MetricListAllDBRepository {
	return &MetricListAllDBRepository{
		db:         db,
		txProvider: txProvider,
	}
}

func (r *MetricListAllDBRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	executor := getExecutor(ctx, r.db, r.txProvider)

	stmt, err := executor.PrepareNamedContext(ctx, metricListAllQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var metrics []types.Metrics
	if err := stmt.SelectContext(ctx, &metrics, map[string]interface{}{}); err != nil {
		return nil, err
	}

	return metrics, nil
}

const metricListAllQuery = `
SELECT id, type, delta, value
FROM content.metrics
ORDER BY id;
`
