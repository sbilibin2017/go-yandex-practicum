package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Generic Tx getter type
type TxGetterFunc func(ctx context.Context) (*sqlx.Tx, bool)

// --- MetricDBSaveRepository ---

type MetricDBSaveRepository struct {
	db       *sqlx.DB
	TxGetter TxGetterFunc
}

type MetricDBSaveRepositoryOption func(*MetricDBSaveRepository)

func WithMetricDBSaveRepositoryDB(db *sqlx.DB) MetricDBSaveRepositoryOption {
	return func(repo *MetricDBSaveRepository) {
		repo.db = db
	}
}

func WithMetricDBSaveRepositoryTxGetter(getter TxGetterFunc) MetricDBSaveRepositoryOption {
	return func(repo *MetricDBSaveRepository) {
		repo.TxGetter = getter
	}
}

func NewMetricDBSaveRepository(opts ...MetricDBSaveRepositoryOption) *MetricDBSaveRepository {
	repo := &MetricDBSaveRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func (r *MetricDBSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	var execer sqlx.ExtContext
	if r.TxGetter != nil {
		if tx, ok := r.TxGetter(ctx); ok && tx != nil {
			execer = tx
		}
	}
	if execer == nil {
		execer = r.db
	}

	_, err := execer.ExecContext(ctx, metricSaveQuery,
		metric.ID,
		metric.Type,
		metric.Delta,
		metric.Value,
	)
	return err
}

const metricSaveQuery = `
INSERT INTO content.metrics (id, type, delta, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, type) DO UPDATE
    SET delta = EXCLUDED.delta,
        value = EXCLUDED.value;
`

// --- MetricDBGetRepository ---

type MetricDBGetRepository struct {
	db       *sqlx.DB
	TxGetter TxGetterFunc
}

type MetricDBGetRepositoryOption func(*MetricDBGetRepository)

func WithMetricDBGetRepositoryDB(db *sqlx.DB) MetricDBGetRepositoryOption {
	return func(repo *MetricDBGetRepository) {
		repo.db = db
	}
}

func WithMetricDBGetRepositoryTxGetter(getter TxGetterFunc) MetricDBGetRepositoryOption {
	return func(repo *MetricDBGetRepository) {
		repo.TxGetter = getter
	}
}

func NewMetricDBGetRepository(opts ...MetricDBGetRepositoryOption) *MetricDBGetRepository {
	repo := &MetricDBGetRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func (r *MetricDBGetRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	var querier sqlx.ExtContext
	if r.TxGetter != nil {
		if tx, ok := r.TxGetter(ctx); ok && tx != nil {
			querier = tx
		}
	}
	if querier == nil {
		querier = r.db
	}

	var metric types.Metrics
	err := sqlx.GetContext(ctx, querier, &metric, metricGetQuery, id.ID, id.Type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &metric, nil
}

const metricGetQuery = `
SELECT id, type, delta, value
FROM content.metrics
WHERE id = $1 AND type = $2;
`

// --- MetricDBListRepository ---

type MetricDBListRepository struct {
	db       *sqlx.DB
	TxGetter TxGetterFunc
}

type MetricDBListRepositoryOption func(*MetricDBListRepository)

func WithMetricDBListRepositoryDB(db *sqlx.DB) MetricDBListRepositoryOption {
	return func(repo *MetricDBListRepository) {
		repo.db = db
	}
}

func WithMetricDBListRepositoryTxGetter(getter TxGetterFunc) MetricDBListRepositoryOption {
	return func(repo *MetricDBListRepository) {
		repo.TxGetter = getter
	}
}

func NewMetricDBListRepository(opts ...MetricDBListRepositoryOption) *MetricDBListRepository {
	repo := &MetricDBListRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func (r *MetricDBListRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	var querier sqlx.ExtContext
	if r.TxGetter != nil {
		if tx, ok := r.TxGetter(ctx); ok && tx != nil {
			querier = tx
		}
	}
	if querier == nil {
		querier = r.db
	}

	var metrics []*types.Metrics
	err := sqlx.SelectContext(ctx, querier, &metrics, metricListQuery)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

const metricListQuery = `
SELECT id, type, delta, value
FROM content.metrics
ORDER BY id;
`
