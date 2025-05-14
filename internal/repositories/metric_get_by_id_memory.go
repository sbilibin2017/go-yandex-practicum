package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDMemoryRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricGetByIDMemoryRepository(
	data map[types.MetricID]types.Metrics,
) *MetricGetByIDMemoryRepository {
	return &MetricGetByIDMemoryRepository{
		data: data,
	}
}

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
