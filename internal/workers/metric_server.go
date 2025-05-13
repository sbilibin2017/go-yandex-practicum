package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricMemoryListAllRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricFileSaveRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricFileListAllRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricMemorySaveRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

func StartMetricServerWorker(
	ctx context.Context,
	metricMemoryListAllRepository MetricMemoryListAllRepository,
	metricMemorySaveRepository MetricMemorySaveRepository,
	metricFileListAllRepository MetricFileListAllRepository,
	metricFileSaveRepository MetricFileSaveRepository,
	storeTicker *time.Ticker,
	restore bool,
) {
	if restore {
		if err := loadMetricsFromFile(ctx, metricFileListAllRepository, metricMemorySaveRepository); err != nil {
			logger.Log.Error("Failed to restore metrics", zap.Error(err))
		}
	}

	if storeTicker == nil {
		<-ctx.Done()
		logger.Log.Info("Context done, saving metrics synchronously (storeInterval = 0)")
		if err := saveMetricsToFile(ctx, metricMemoryListAllRepository, metricFileSaveRepository); err != nil {
			logger.Log.Error("Error saving metrics before shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Metrics saved before shutdown")
		}
		return
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Shutting down server...")
			saveMetricsToFile(ctx, metricMemoryListAllRepository, metricFileSaveRepository)
			return
		case <-storeTicker.C:
			saveMetricsToFile(ctx, metricMemoryListAllRepository, metricFileSaveRepository)
		}
	}
}

func saveMetricsToFile(
	ctx context.Context,
	metricMemoryListAllRepository MetricMemoryListAllRepository,
	metricFileSaveRepository MetricFileSaveRepository,
) error {
	logger.Log.Info("Starting periodic/synchronous save")
	metrics, err := metricMemoryListAllRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from memory", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		logger.Log.Sugar().Infof("Saving metric: %+v", metric)
		if err := metricFileSaveRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to file", zap.Error(err))
			return err
		}
	}
	logger.Log.Info("Metrics saved to file successfully")
	return nil
}

func loadMetricsFromFile(
	ctx context.Context,
	metricFileListAllRepository MetricFileListAllRepository,
	metricMemorySaveRepository MetricMemorySaveRepository,
) error {
	metrics, err := metricFileListAllRepository.ListAll(ctx)
	if err != nil {
		logger.Log.Error("Error fetching metrics from file", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		if err := metricMemorySaveRepository.Save(ctx, metric); err != nil {
			logger.Log.Error("Error saving metric to memory", zap.Error(err))
			return err
		}
	}
	logger.Log.Info("Metrics loaded from file and saved in memory successfully")
	return nil
}
