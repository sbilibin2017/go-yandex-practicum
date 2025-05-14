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
	metrics, err := namedQueryContext[types.Metrics](ctx, r.db, r.txProvider, metricListAllQuery, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

const metricListAllQuery = `
SELECT id, type, delta, value
FROM content.metrics
ORDER BY id;
`
