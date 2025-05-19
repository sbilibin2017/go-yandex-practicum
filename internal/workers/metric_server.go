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
