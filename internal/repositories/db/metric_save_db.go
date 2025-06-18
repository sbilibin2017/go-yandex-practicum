package db

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaveDBRepository реализует сохранение метрик в базу данных.
type MetricSaveDBRepository struct {
	db         *sqlx.DB
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor
}

// NewMetricSaveDBRepository создает новый репозиторий для сохранения метрик в БД.
//
// Параметры:
//   - db: подключение к базе данных sqlx.DB.
//   - execGetter: функция для получения исполнителя (executor) из контекста.
//
// Возвращает:
//   - указатель на MetricSaveDBRepository.
func NewMetricSaveDBRepository(
	db *sqlx.DB,
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor,
) *MetricSaveDBRepository {
	return &MetricSaveDBRepository{
		db:         db,
		execGetter: execGetter,
	}
}

// Save сохраняет метрику в базу данных.
//
// Если метрика с таким же id и типом уже существует, данные обновляются.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - metric: структура метрики для сохранения.
//
// Возвращает:
//   - ошибку, если сохранение прошло неуспешно.
func (r *MetricSaveDBRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	args := map[string]any{
		"id":    metric.ID,
		"type":  metric.Type,
		"delta": metric.Delta,
		"value": metric.Value,
	}

	executor := r.execGetter(ctx, r.db)

	stmt, err := executor.PrepareNamedContext(ctx, metricSaveQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args)
	return err
}

const metricSaveQuery = `
INSERT INTO content.metrics (id, type, delta, value)
VALUES (:id, :type, :delta, :value)
ON CONFLICT (id, type) DO UPDATE
    SET delta = EXCLUDED.delta,
        value = EXCLUDED.value;
`
