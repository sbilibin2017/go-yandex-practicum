package repositories

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricGetByIDFileRepository реализует репозиторий для чтения метрик из файла.
//
// Хранилище работает с JSON-данными, представляющими собой последовательность
// закодированных объектов типа Metrics, и обеспечивает потокобезопасный доступ.
type MetricGetByIDFileRepository struct {
	file *os.File
	mu   sync.Mutex
}

// NewMetricGetByIDFileRepository создает новый экземпляр MetricGetByIDFileRepository.
//
// Параметры:
//   - file: дескриптор открытого файла, содержащего сериализованные метрики.
//
// Возвращает:
//   - *MetricGetByIDFileRepository: экземпляр репозитория.
func NewMetricGetByIDFileRepository(file *os.File) *MetricGetByIDFileRepository {
	return &MetricGetByIDFileRepository{
		file: file,
	}
}

// GetByID ищет метрику по заданному идентификатору и типу в файле.
//
// Последовательно декодирует объекты JSON из файла и сравнивает их с переданным идентификатором.
// При нахождении соответствующей метрики возвращает её. Если ни одна метрика не найдена — возвращает nil.
// Файл синхронизируется до и после чтения, а доступ к нему защищён мьютексом.
//
// Параметры:
//   - ctx: контекст выполнения.
//   - id: структура идентификатора метрики (тип и имя).
//
// Возвращает:
//   - *types.Metrics: найденная метрика или nil.
//   - error: ошибка, если чтение файла не удалось.
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
