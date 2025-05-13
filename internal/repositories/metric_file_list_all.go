package repositories

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFileListAllRepository struct {
	decoder *json.Decoder
	file    *os.File
	mu      sync.Mutex
}

func NewMetricFileListAllRepository(file *os.File) *MetricFileListAllRepository {
	reader := bufio.NewReader(file)
	return &MetricFileListAllRepository{
		file:    file,
		decoder: json.NewDecoder(reader),
	}
}

func (r *MetricFileListAllRepository) ListAll(
	ctx context.Context,
) ([]types.Metrics, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := r.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(r.file)
	decoder := json.NewDecoder(reader)

	metrics := make(map[types.MetricID]types.Metrics)
	for {
		var metric types.Metrics
		err := decoder.Decode(&metric)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		metrics[metric.MetricID] = metric
	}

	var metricsSlice []types.Metrics
	for _, m := range metrics {
		metricsSlice = append(metricsSlice, m)
	}

	sort.SliceStable(metricsSlice, func(i, j int) bool {
		return metricsSlice[i].ID < metricsSlice[j].ID
	})

	return metricsSlice, nil
}
