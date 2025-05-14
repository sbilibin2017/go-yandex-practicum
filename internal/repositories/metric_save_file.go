package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveFileRepository struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

func NewMetricSaveFileRepository(file *os.File) *MetricSaveFileRepository {
	return &MetricSaveFileRepository{
		file:    file,
		encoder: json.NewEncoder(file),
	}
}

func (r *MetricSaveFileRepository) Save(ctx context.Context, metric types.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return withFileSync(r.file, func(f *os.File) error {
		encoder := json.NewEncoder(f)
		return encoder.Encode(metric)
	})
}
