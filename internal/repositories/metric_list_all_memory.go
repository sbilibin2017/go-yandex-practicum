package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllMemoryRepository реализует репозиторий для получения всех метрик из памяти.
type MetricListAllMemoryRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

// NewMetricListAllMemoryRepository создает новый экземпляр MetricListAllMemoryRepository.
//
// Параметры:
//   - data: карта метрик с ключом типа MetricID и значением типа Metrics.
//
// Возвращает:
//   - указатель на созданный репозиторий.
func NewMetricListAllMemoryRepository(
	data map[types.MetricID]types.Metrics,
) *MetricListAllMemoryRepository {
	return &MetricListAllMemoryRepository{
		data: data,
	}
}

// ListAll возвращает срез всех метрик из памяти, отсортированных по ID.
//
// Параметры:
//   - ctx: контекст выполнения.
//
// Возвращает:
//   - срез метрик ([]types.Metrics), отсортированный по ID.
//   - ошибку (в данном случае всегда nil).
func (r *MetricListAllMemoryRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var metrics []types.Metrics
	for _, m := range r.data {
		metrics = append(metrics, m)
	}

	// Сортируем метрики по ID в стабильном порядке
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}
