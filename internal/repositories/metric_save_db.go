package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaveDBRepository реализует сохранение метрик в базу данных.
type MetricSaveDBRepository struct {
	db         *sqlx.DB
	txProvider func(ctx context.Context) *sqlx.Tx
}

// NewMetricSaveDBRepository создает новый репозиторий для сохранения метрик в БД.
//
// Параметры:
//   - db: подключение к базе данных sqlx.DB.
//   - txProvider: функция для получения транзакции из контекста.
//
// Возвращает:
//   - указатель на MetricSaveDBRepository.
func NewMetricSaveDBRepository(
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
) *MetricSaveDBRepository {
	return &MetricSaveDBRepository{
		db:         db,
		txProvider: txProvider,
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

	executor := getExecutor(ctx, r.db, r.txProvider)

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
