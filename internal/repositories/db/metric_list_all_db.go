package db

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllDBRepository реализует репозиторий для получения всех метрик из базы данных.
type MetricListAllDBRepository struct {
	db         *sqlx.DB
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor
}

// NewMetricListAllDBRepository создает новый экземпляр MetricListAllDBRepository.
//
// Параметры:
//   - db: экземпляр базы данных sqlx.DB.
//   - execGetter: функция для получения исполнителя (executor) из контекста.
//
// Возвращает:
//   - указатель на созданный репозиторий.
func NewMetricListAllDBRepository(
	db *sqlx.DB,
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor,
) *MetricListAllDBRepository {
	return &MetricListAllDBRepository{
		db:         db,
		execGetter: execGetter,
	}
}

// ListAll возвращает список всех метрик из базы данных.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//
// Возвращает:
//   - срез метрик ([]types.Metrics).
//   - ошибку в случае неудачи при выполнении запроса.
func (r *MetricListAllDBRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	executor := r.execGetter(ctx, r.db)

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
