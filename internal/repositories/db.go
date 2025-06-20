package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveDBRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricSaveDBRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricSaveDBRepository {
	return &MetricSaveDBRepository{
		db:       db,
		txGetter: txGetter,
	}
}

func (r *MetricSaveDBRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	args := map[string]any{
		"id":    metric.ID,
		"type":  metric.MType,
		"delta": metric.Delta,
		"value": metric.Value,
	}

	var executor sqlx.ExtContext
	if tx := r.txGetter(ctx); tx != nil {
		executor = tx
	} else {
		executor = r.db
	}

	_, err := sqlx.NamedExecContext(ctx, executor, metricSaveQuery, args)
	return err
}

const metricSaveQuery = `
INSERT INTO content.metrics (id, type, delta, value)
VALUES (:id, :type, :delta, :value)
ON CONFLICT (id, type) DO UPDATE
SET delta = EXCLUDED.delta,
    value = EXCLUDED.value;
`

type MetricGetByIDDBRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricGetByIDDBRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricGetByIDDBRepository {
	return &MetricGetByIDDBRepository{db: db, txGetter: txGetter}
}

func (r *MetricGetByIDDBRepository) Get(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	args := map[string]any{
		"id":   id.ID,
		"type": id.MType,
	}

	var executor sqlx.ExtContext
	if tx := r.txGetter(ctx); tx != nil {
		executor = tx
	} else {
		executor = r.db
	}

	var metric types.Metrics
	rows, err := sqlx.NamedQueryContext(ctx, executor, metricGetByIDQuery, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&metric); err != nil {
			return nil, err
		}
		return &metric, nil
	}

	return nil, nil
}

const metricGetByIDQuery = `
SELECT id, type, delta, value
FROM content.metrics
WHERE id = :id AND type = :type;
`

type MetricListAllDBRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricListAllDBRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricListAllDBRepository {
	return &MetricListAllDBRepository{
		db:       db,
		txGetter: txGetter,
	}
}

func (r *MetricListAllDBRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	var executor sqlx.ExtContext
	if tx := r.txGetter(ctx); tx != nil {
		executor = tx
	} else {
		executor = r.db
	}

	rows, err := sqlx.NamedQueryContext(ctx, executor, metricListAllQuery, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*types.Metrics
	for rows.Next() {
		var metric types.Metrics
		if err := rows.StructScan(&metric); err != nil {
			return nil, err
		}
		metrics = append(metrics, &metric)
	}

	return metrics, nil
}

const metricListAllQuery = `
SELECT id, type, delta, value
FROM content.metrics
ORDER BY id;
`
