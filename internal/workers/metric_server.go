package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricSaveRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricListAllRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricSaveFileRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricListAllFileRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricServerWorker struct {
	memoryListAllRepo MetricListAllRepository
	memorySaveRepo    MetricSaveRepository
	fileListAllRepo   MetricListAllFileRepository
	fileSaveRepo      MetricSaveFileRepository
	storeTicker       *time.Ticker
	restore           bool
}

func NewMetricServerWorker(
	memoryListAll MetricListAllRepository,
	memorySave MetricSaveRepository,
	fileListAll MetricListAllFileRepository,
	fileSave MetricSaveFileRepository,
	storeTicker *time.Ticker,
	restore bool,
) *MetricServerWorker {
	return &MetricServerWorker{
		memoryListAllRepo: memoryListAll,
		memorySaveRepo:    memorySave,
		fileListAllRepo:   fileListAll,
		fileSaveRepo:      fileSave,
		storeTicker:       storeTicker,
		restore:           restore,
	}
}

func (w *MetricServerWorker) Start(ctx context.Context) error {
	if w.restore {
		if err := loadMetricsFromFile(ctx, w.fileListAllRepo, w.memorySaveRepo); err != nil {
			logger.Log.Error("Failed to restore metrics", zap.Error(err))
		}
	}

	if w.storeTicker == nil {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := saveMetricsToFile(shutdownCtx, w.memoryListAllRepo, w.fileSaveRepo); err != nil {
			logger.Log.Error("Error saving metrics before shutdown", zap.Error(err))
			return err
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return saveMetricsToFile(shutdownCtx, w.memoryListAllRepo, w.fileSaveRepo)

		case <-w.storeTicker.C:
			if err := saveMetricsToFile(ctx, w.memoryListAllRepo, w.fileSaveRepo); err != nil {
				logger.Log.Error("Periodic save failed", zap.Error(err))
			}
		}
	}
}

func saveMetricsToFile(
	ctx context.Context,
	MetricListAllMemoryRepository MetricListAllRepository,
	MetricSaveFileRepository MetricSaveFileRepository,
) error {
	metrics, err := MetricListAllMemoryRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from memory", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		if err := MetricSaveFileRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to file", zap.Error(err))
			return err
		}
	}
	return nil
}

func loadMetricsFromFile(
	ctx context.Context,
	MetricListAllFileRepository MetricListAllFileRepository,
	MetricSaveMemoryRepository MetricSaveRepository,
) error {
	metrics, err := MetricListAllFileRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from file", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		if err := MetricSaveMemoryRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to memory", zap.Error(err))
			return err
		}
	}
	return nil
}
