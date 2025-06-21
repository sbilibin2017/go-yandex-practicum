package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBConfig holds the DB instance and optional transaction getter function.
type DBConfig struct {
	DB       *sqlx.DB
	TxGetter func(ctx context.Context) (*sqlx.Tx, bool)
}

// DBOption configures DBConfig.
type DBOption func(*DBConfig)

// NewDBConfig creates a new DBConfig applying the given options.
func NewDBConfig(opts ...DBOption) *DBConfig {
	cfg := &DBConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithDB sets the DB instance in the DBConfig.
func WithDB(db *sqlx.DB) DBOption {
	return func(cfg *DBConfig) {
		cfg.DB = db
	}
}

// WithTxGetter sets the transaction getter function in the DBConfig.
func WithTxGetter(opt func(ctx context.Context) (*sqlx.Tx, bool)) DBOption {
	return func(cfg *DBConfig) {
		cfg.TxGetter = opt
	}
}

// MetricDBSaveRepository handles saving metrics to the database.
type MetricDBSaveRepository struct {
	config *DBConfig
}

// NewMetricDBSaveRepository creates a new MetricDBSaveRepository with the given options.
func NewMetricDBSaveRepository(opts ...DBOption) *MetricDBSaveRepository {
	return &MetricDBSaveRepository{config: NewDBConfig(opts...)}
}

// Save inserts or updates a metric in the database.
// Uses transaction from context if available; otherwise, uses the DB connection.
func (r *MetricDBSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	tx, ok := r.config.TxGetter(ctx)

	var execer sqlx.ExtContext
	if ok && tx != nil {
		execer = tx
	} else {
		execer = r.config.DB
	}

	_, err := execer.ExecContext(ctx, metricSaveQuery,
		metric.ID,
		metric.MType,
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

// MetricDBGetRepository handles retrieving metrics from the database.
type MetricDBGetRepository struct {
	config *DBConfig
}

// NewMetricDBGetRepository creates a new MetricDBGetRepository with the given options.
func NewMetricDBGetRepository(opts ...DBOption) *MetricDBGetRepository {
	return &MetricDBGetRepository{config: NewDBConfig(opts...)}
}

// Get retrieves a metric by ID and type from the database.
// Returns (nil, nil) if no matching metric is found.
// Uses transaction from context if available; otherwise, uses the DB connection.
func (r *MetricDBGetRepository) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	tx, ok := r.config.TxGetter(ctx)

	var querier sqlx.ExtContext
	if ok && tx != nil {
		querier = tx
	} else {
		querier = r.config.DB
	}

	var metric types.Metrics
	err := sqlx.GetContext(ctx, querier, &metric, metricGetQuery, id.ID, id.MType)
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

// MetricDBListRepository handles listing all metrics from the database.
type MetricDBListRepository struct {
	config *DBConfig
}

// NewMetricDBListRepository creates a new MetricDBListRepository with the given options.
func NewMetricDBListRepository(opts ...DBOption) *MetricDBListRepository {
	return &MetricDBListRepository{config: NewDBConfig(opts...)}
}

// List returns all metrics from the database ordered by ID.
// Uses transaction from context if available; otherwise, uses the DB connection.
func (r *MetricDBListRepository) List(ctx context.Context) ([]*types.Metrics, error) {
	tx, ok := r.config.TxGetter(ctx)

	var querier sqlx.ExtContext
	if ok && tx != nil {
		querier = tx
	} else {
		querier = r.config.DB
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
