package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricMemoryGetByIDRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemoryGetByIDRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemoryGetByIDRepository {
	return &MetricMemoryGetByIDRepository{
		data: data,
	}
}

func (r *MetricMemoryGetByIDRepository) GetByID(
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
