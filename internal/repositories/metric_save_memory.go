package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaveMemoryRepository реализует сохранение метрик в памяти.
type MetricSaveMemoryRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

// NewMetricSaveMemoryRepository создает новый репозиторий для сохранения метрик в памяти.
//
// Параметры:
//   - data: карта метрик, используемая для хранения.
//
// Возвращает:
//   - указатель на MetricSaveMemoryRepository.
func NewMetricSaveMemoryRepository(
	data map[types.MetricID]types.Metrics,
) *MetricSaveMemoryRepository {
	return &MetricSaveMemoryRepository{
		data: data,
	}
}

// Save сохраняет метрику в памяти.
//
// Метод блокирует запись для обеспечения безопасности при конкурентном доступе.
//
// Параметры:
//   - ctx: контекст выполнения операции.
//   - metric: структура метрики для сохранения.
//
// Возвращает:
//   - ошибку, которая всегда nil, так как сохранение в памяти не может завершиться с ошибкой.
func (r *MetricSaveMemoryRepository) Save(
	ctx context.Context, metric types.Metrics,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[metric.MetricID] = metric
	return nil
}
