package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

// MetricSaveRepository описывает интерфейс для сохранения метрики в память или базу данных.
type MetricSaveRepository interface {
	// Save сохраняет переданную метрику.
	Save(ctx context.Context, metric types.Metrics) error
}

// MetricListAllRepository описывает интерфейс для получения всех метрик из памяти или базы данных.
type MetricListAllRepository interface {
	// ListAll возвращает список всех метрик.
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

// MetricSaveFileRepository описывает интерфейс для сохранения метрики в файл.
type MetricSaveFileRepository interface {
	// Save сохраняет метрику в файл.
	Save(ctx context.Context, metric types.Metrics) error
}

// MetricListAllFileRepository описывает интерфейс для получения всех метрик из файла.
type MetricListAllFileRepository interface {
	// ListAll возвращает список всех метрик из файла.
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

// NewMetricServerWorker создаёт функцию-воркер, которая управляет
// периодическим сохранением метрик из памяти в файл и восстановлением метрик из файла.
// Параметры:
//   - ctx: контекст для отмены работы воркера,
//   - memoryListAll: репозиторий для чтения всех метрик из памяти,
//   - memorySave: репозиторий для сохранения метрик в память,
//   - fileListAll: репозиторий для чтения метрик из файла,
//   - fileSave: репозиторий для сохранения метрик в файл,
//   - restore: флаг, указывающий, нужно ли восстанавливать метрики из файла при старте,
//   - storeInterval: интервал в секундах для периодического сохранения метрик в файл.
//
// Если storeInterval равен 0, сохранение происходит только при завершении работы.
func NewMetricServerWorker(
	ctx context.Context,
	memoryListAll MetricListAllRepository,
	memorySave MetricSaveRepository,
	fileListAll MetricListAllFileRepository,
	fileSave MetricSaveFileRepository,
	restore bool,
	storeInterval int,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return startMetricServerWorker(
			ctx,
			memoryListAll,
			memorySave,
			fileListAll,
			fileSave,
			restore,
			storeInterval,
		)
	}
}

// startMetricServerWorker запускает основной цикл работы воркера, который
// восстанавливает метрики из файла, а затем периодически сохраняет их обратно.
func startMetricServerWorker(
	ctx context.Context,
	memoryListAll MetricListAllRepository,
	memorySave MetricSaveRepository,
	fileListAll MetricListAllFileRepository,
	fileSave MetricSaveFileRepository,
	restore bool,
	storeInterval int,
) error {
	if restore {
		if err := loadMetricsFromFile(ctx, fileListAll, memorySave); err != nil {
			logger.Log.Error("Failed to restore metrics", zap.Error(err))
			return err
		}
	}

	if storeInterval == 0 {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := saveMetricsToFile(shutdownCtx, memoryListAll, fileSave); err != nil {
			logger.Log.Error("Error saving metrics before shutdown", zap.Error(err))
			return err
		}
		return nil
	}

	storeTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return saveMetricsToFile(shutdownCtx, memoryListAll, fileSave)
		case <-storeTicker.C:
			err := saveMetricsToFile(ctx, memoryListAll, fileSave)
			if err != nil {
				return err
			}
		}
	}
}

// saveMetricsToFile сохраняет все метрики из памяти в файл.
// Возвращает ошибку в случае неудачи.
func saveMetricsToFile(
	ctx context.Context,
	metricListAllMemoryRepository MetricListAllRepository,
	metricSaveFileRepository MetricSaveFileRepository,
) error {
	metrics, err := metricListAllMemoryRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from memory", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		if err := metricSaveFileRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to file", zap.Error(err))
			return err
		}
	}
	return nil
}

// loadMetricsFromFile загружает все метрики из файла и сохраняет их в память.
// Возвращает ошибку в случае неудачи.
func loadMetricsFromFile(
	ctx context.Context,
	metricListAllFileRepository MetricListAllFileRepository,
	metricSaveMemoryRepository MetricSaveRepository,
) error {
	metrics, err := metricListAllFileRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from file", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		if err := metricSaveMemoryRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to memory", zap.Error(err))
			return err
		}
	}
	return nil
}
