package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetByIDMemoryRepository реализует репозиторий для чтения метрик из памяти.
//
// Обеспечивает потокобезопасный доступ к метрикам, хранящимся в памяти в виде карты.
type MetricGetByIDMemoryRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

// NewMetricGetByIDMemoryRepository создает новый экземпляр MetricGetByIDMemoryRepository.
//
// Параметры:
//   - data: карта метрик, где ключом является MetricID.
//
// Возвращает:
//   - *MetricGetByIDMemoryRepository: новый репозиторий для работы с метриками.
func NewMetricGetByIDMemoryRepository(
	data map[types.MetricID]types.Metrics,
) *MetricGetByIDMemoryRepository {
	return &MetricGetByIDMemoryRepository{
		data: data,
	}
}

// GetByID возвращает метрику по заданному идентификатору.
//
// Если метрика с указанным id существует, возвращает указатель на нее,
// иначе — возвращает nil без ошибки.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - id: идентификатор метрики (ID и тип).
//
// Возвращает:
//   - *types.Metrics: найденная метрика или nil, если не найдена.
//   - error: всегда nil (ошибки не возникают).
func (r *MetricGetByIDMemoryRepository) GetByID(
	ctx context.Context, id types.MetricID,
) (*types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.data[id]
	if !exists {
		return nil, nil
	}
	return &metric, nil
}
