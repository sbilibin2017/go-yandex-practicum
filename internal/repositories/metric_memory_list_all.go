package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricMemoryListAllRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemoryListAllRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemoryListAllRepository {
	return &MetricMemoryListAllRepository{
		data: data,
	}
}

func (r *MetricMemoryListAllRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var metrics []types.Metrics
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics, nil
}
