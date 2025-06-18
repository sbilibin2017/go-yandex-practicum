package memory

import (
	"context"
	"sort"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricListAllMemoryRepository реализует репозиторий для получения всех метрик из памяти.
type MetricListAllMemoryRepository struct{}

// NewMetricListAllMemoryRepository создает новый экземпляр MetricListAllMemoryRepository.
func NewMetricListAllMemoryRepository() *MetricListAllMemoryRepository {
	return &MetricListAllMemoryRepository{}
}

// ListAll возвращает срез всех метрик из памяти, отсортированных по ID.
func (r *MetricListAllMemoryRepository) ListAll(ctx context.Context) ([]types.Metrics, error) {
	mu.RLock()
	defer mu.RUnlock()

	var metrics []types.Metrics
	for _, m := range data {
		metrics = append(metrics, m)
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}
