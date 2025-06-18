package db

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetByIDDBRepository представляет реализацию репозитория для получения метрики по идентификатору из базы данных.
type MetricGetByIDDBRepository struct {
	db         *sqlx.DB
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor
}

// NewMetricGetByIDDBRepository создает новый экземпляр репозитория MetricGetByIDDBRepository.
// Параметры:
//   - db: подключение к базе данных.
//   - txGetter: функция получения текущей транзакции из контекста.
func NewMetricGetByIDDBRepository(
	db *sqlx.DB,
	execGetter func(ctx context.Context, db *sqlx.DB) middlewares.Executor,
) *MetricGetByIDDBRepository {
	return &MetricGetByIDDBRepository{db: db, execGetter: execGetter}
}

// GetByID извлекает метрику из базы данных по идентификатору (id и type).
// Использует транзакцию из контекста, если она присутствует.
// Параметры:
//   - ctx: контекст выполнения.
//   - id: структура идентификатора метрики.
//
// Возвращает найденную метрику или ошибку, если метрика не найдена или возникла проблема с запросом.
func (r *MetricGetByIDDBRepository) GetByID(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	args := map[string]any{
		"id":   id.ID,
		"type": id.Type,
	}

	var metric types.Metrics

	executor := r.execGetter(ctx, r.db)

	stmt, err := executor.PrepareNamedContext(ctx, metricGetByIDQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &metric, args); err != nil {
		return nil, err
	}

	return &metric, nil
}

// SQL-запрос для получения метрики по id и type.
const metricGetByIDQuery = `
SELECT id, type, delta, value
FROM content.metrics
WHERE id = :id AND type = :type;
`
