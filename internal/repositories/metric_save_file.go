package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricSaveFileRepository реализует сохранение метрик в файл в формате JSON.
type MetricSaveFileRepository struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

// NewMetricSaveFileRepository создает новый репозиторий для сохранения метрик в файл.
//
// Параметры:
//   - file: указатель на файл для записи метрик.
//
// Возвращает:
//   - указатель на MetricSaveFileRepository.
func NewMetricSaveFileRepository(file *os.File) *MetricSaveFileRepository {
	return &MetricSaveFileRepository{
		file:    file,
		encoder: json.NewEncoder(file),
	}
}

// Save сохраняет метрику в файл.
//
// Метод блокирует доступ к файлу, чтобы избежать состояния гонки при параллельной записи.
// Сохраняет метрику в JSON формате, записывая ее в файл.
//
// Параметры:
//   - ctx: контекст выполнения операции.
//   - metric: структура метрики для сохранения.
//
// Возвращает:
//   - ошибку, если не удалось записать метрику в файл.
func (r *MetricSaveFileRepository) Save(ctx context.Context, metric types.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return withFileSync(r.file, func(f *os.File) error {
		encoder := json.NewEncoder(f)
		return encoder.Encode(metric)
	})
}
