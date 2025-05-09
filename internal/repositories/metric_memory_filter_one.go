package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFilterOneRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricFilterOneRepository(
	data map[types.MetricID]types.Metrics,
) *MetricFilterOneRepository {
	return &MetricFilterOneRepository{
		data: data,
	}
}

func (r *MetricFilterOneRepository) FilterOne(
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
