package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveDBRepository struct {
	db         *sqlx.DB
	txProvider func(ctx context.Context) *sqlx.Tx
}

func NewMetricSaveDBRepository(
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
) *MetricSaveDBRepository {
	return &MetricSaveDBRepository{
		db:         db,
		txProvider: txProvider,
	}
}

func (r *MetricSaveDBRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	args := map[string]any{
		"id":    metric.ID,
		"type":  metric.Type,
		"delta": metric.Delta,
		"value": metric.Value,
	}
	return namedExecContext(ctx, r.db, r.txProvider, metricSaveQuery, args)
}

const metricSaveQuery = `
INSERT INTO content.metrics (id, type, delta, value)
VALUES (:id, :type, :delta, :value)
ON CONFLICT (id, type) DO UPDATE
    SET delta = EXCLUDED.delta,
        value = EXCLUDED.value;
`
