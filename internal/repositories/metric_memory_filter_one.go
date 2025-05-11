package repositories

import (
	"context"
	"sync"
)

type MetricFilterOneRepository struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMetricFilterOneRepository(
	data map[string]any,
) *MetricFilterOneRepository {
	return &MetricFilterOneRepository{
		data: data,
	}
}

func (r *MetricFilterOneRepository) FilterOne(
	ctx context.Context, filter map[string]any,
) (map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.data[generateMetricKey(filter)]
	if !exists {
		return nil, nil
	}
	return metric.(map[string]any), nil
}
