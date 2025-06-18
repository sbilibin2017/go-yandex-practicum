package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricSaveFileRepository struct {
	pathToFile string
}

func NewMetricSaveFileRepository(pathToFile string) *MetricSaveFileRepository {
	return &MetricSaveFileRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricSaveFileRepository) Save(ctx context.Context, metric types.Metrics) error {
	mu.Lock()
	defer mu.Unlock()

	origFile, err := os.Open(r.pathToFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if origFile != nil {
		defer origFile.Close()
	}

	tmpFilePath := r.pathToFile + ".tmp"
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	decoder := json.NewDecoder(origFile)
	encoder := json.NewEncoder(tmpFile)

	found := false

	for {
		var m types.Metrics
		err := decoder.Decode(&m)
		if err != nil {
			break
		}

		if m.ID == metric.ID && m.Type == metric.Type {
			if err := encoder.Encode(metric); err != nil {
				return err
			}
			found = true
		} else {
			if err := encoder.Encode(m); err != nil {
				return err
			}
		}
	}

	if !found {
		if err := encoder.Encode(metric); err != nil {
			return err
		}
	}

	if err := os.Rename(tmpFilePath, r.pathToFile); err != nil {
		return err
	}

	return nil
}
