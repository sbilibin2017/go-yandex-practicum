package repositories

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFileSaveRepository struct {
	encoder *json.Encoder
	writer  *bufio.Writer
	mu      sync.Mutex
}

func NewMetricFileSaveRepository(file *os.File) *MetricFileSaveRepository {
	writer := bufio.NewWriter(file)
	return &MetricFileSaveRepository{
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}
}

func (r *MetricFileSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	log.Println("Starting to save metric")
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.encoder.Encode(metric); err != nil {
		return err
	}
	if err := r.writer.Flush(); err != nil {
		return err
	}
	return nil
}
