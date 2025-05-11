package repositories

import (
	"context"
	"sort"
	"sync"
)

type MetricListAllRepository struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMetricListAllRepository(
	data map[string]any,
) *MetricListAllRepository {
	return &MetricListAllRepository{
		data: data,
	}
}

func (r *MetricListAllRepository) ListAll(
	ctx context.Context,
) ([]map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []map[string]any
	for _, v := range r.data {
		if item, ok := v.(map[string]any); ok {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		id1, ok1 := result[i]["id"].(string)
		id2, ok2 := result[j]["id"].(string)
		if !ok1 || !ok2 {
			return false
		}
		return id1 < id2
	})

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}
