package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricListAllMemoryRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricSaveFileRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricListAllFileRepository interface {
	ListAll(ctx context.Context) ([]types.Metrics, error)
}

type MetricSaveMemoryRepository interface {
	Save(ctx context.Context, metric types.Metrics) error
}

func StartMetricServerWorker(
	ctx context.Context,
	MetricListAllMemoryRepository MetricListAllMemoryRepository,
	MetricSaveMemoryRepository MetricSaveMemoryRepository,
	MetricListAllFileRepository MetricListAllFileRepository,
	MetricSaveFileRepository MetricSaveFileRepository,
	storeTicker *time.Ticker,
	restore bool,
) error {
	if restore {
		if err := loadMetricsFromFile(ctx, MetricListAllFileRepository, MetricSaveMemoryRepository); err != nil {
			logger.Log.Error("Failed to restore metrics", zap.Error(err))
		}
	}

	if storeTicker == nil {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := saveMetricsToFile(shutdownCtx, MetricListAllMemoryRepository, MetricSaveFileRepository); err != nil {
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
			return saveMetricsToFile(shutdownCtx, MetricListAllMemoryRepository, MetricSaveFileRepository)

		case <-storeTicker.C:
			if err := saveMetricsToFile(ctx, MetricListAllMemoryRepository, MetricSaveFileRepository); err != nil {
				logger.Log.Error("Periodic save failed", zap.Error(err))
			}
		}
	}
}

func saveMetricsToFile(
	ctx context.Context,
	MetricListAllMemoryRepository MetricListAllMemoryRepository,
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
	MetricSaveMemoryRepository MetricSaveMemoryRepository,
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
