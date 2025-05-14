package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllFileRepository struct {
	file *os.File
	mu   sync.Mutex
}

func NewMetricListAllFileRepository(file *os.File) *MetricListAllFileRepository {
	return &MetricListAllFileRepository{
		file: file,
	}
}

func (r *MetricListAllFileRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var metricsSlice []types.Metrics
	err := withFileSync(r.file, func(f *os.File) error {
		decoder := json.NewDecoder(f)
		metricsMap := make(map[types.MetricID]types.Metrics)
		for {
			var metric types.Metrics
			if err := decoder.Decode(&metric); err != nil {
				break
			}
			metricsMap[metric.MetricID] = metric
		}
		for _, m := range metricsMap {
			metricsSlice = append(metricsSlice, m)
		}
		sort.SliceStable(metricsSlice, func(i, j int) bool {
			return metricsSlice[i].ID < metricsSlice[j].ID
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metricsSlice, nil
}
