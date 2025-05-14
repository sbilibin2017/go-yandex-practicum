package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDDBRepository struct {
	db         *sqlx.DB
	txProvider func(ctx context.Context) *sqlx.Tx
}

func NewMetricGetByIDDBRepository(
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
) *MetricGetByIDDBRepository {
	return &MetricGetByIDDBRepository{
		db:         db,
		txProvider: txProvider,
	}
}

func (r *MetricGetByIDDBRepository) GetByID(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	args := map[string]any{
		"id":   id.ID,
		"type": id.Type,
	}
	metric, err := namedQueryOneContext[types.Metrics](ctx, r.db, r.txProvider, metricGetByIDQuery, args)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

const metricGetByIDQuery = `
SELECT id, type, delta, value
FROM content.metrics
WHERE id = :id AND type = :type;
`
