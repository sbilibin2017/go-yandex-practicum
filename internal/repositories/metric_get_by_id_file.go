package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricGetByIDFileRepository struct {
	file *os.File
	mu   sync.Mutex
}

func NewMetricGetByIDFileRepository(file *os.File) *MetricGetByIDFileRepository {
	return &MetricGetByIDFileRepository{
		file: file,
	}
}

func (r *MetricGetByIDFileRepository) GetByID(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var metricFound *types.Metrics
	err := withFileSync(r.file, func(f *os.File) error {
		decoder := json.NewDecoder(f)
		for {
			var metric types.Metrics
			if err := decoder.Decode(&metric); err != nil {
				break
			}
			if metric.MetricID == id {
				metricFound = &metric
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metricFound, nil
}
