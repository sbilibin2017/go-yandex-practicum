package repositories

import (
	"context"
	"sync"
)

type MetricMemorySaveRepository struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMetricMemorySaveRepository(
	data map[string]any,
) *MetricMemorySaveRepository {
	return &MetricMemorySaveRepository{
		data: data,
	}
}

func (r *MetricMemorySaveRepository) Save(
	ctx context.Context, data map[string]any,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[generateMetricKey(data)] = data
	return nil
}
