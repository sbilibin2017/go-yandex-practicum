package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type Saver interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type Lister interface {
	List(ctx context.Context) ([]*types.Metrics, error)
}

// ServerWorkerOption configures the server worker.
type ServerWorkerOption func(*workerOptions)

type workerOptions struct {
	restore       bool
	storeInterval int
	lister        Lister
	saver         Saver
	listerFile    Lister
	saverFile     Saver
}

func NewServerWorker(opts ...ServerWorkerOption) func(ctx context.Context) error {
	// defaults
	var wo workerOptions

	for _, opt := range opts {
		opt(&wo)
	}

	return func(ctx context.Context) error {
		logger.Log.Debug("ServerWorker: starting")
		return startMetricServerWorker(
			ctx,
			wo.lister,
			wo.saver,
			wo.listerFile,
			wo.saverFile,
			wo.restore,
			wo.storeInterval,
		)
	}
}

// Functional option implementations:

func WithRestore(restore bool) ServerWorkerOption {
	return func(o *workerOptions) {
		o.restore = restore
	}
}

func WithStoreInterval(interval int) ServerWorkerOption {
	return func(o *workerOptions) {
		o.storeInterval = interval
	}
}

func WithLister(lister Lister) ServerWorkerOption {
	return func(o *workerOptions) {
		o.lister = lister
	}
}

func WithSaver(saver Saver) ServerWorkerOption {
	return func(o *workerOptions) {
		o.saver = saver
	}
}

func WithListerFile(listerFile Lister) ServerWorkerOption {
	return func(o *workerOptions) {
		o.listerFile = listerFile
	}
}

func WithSaverFile(saverFile Saver) ServerWorkerOption {
	return func(o *workerOptions) {
		o.saverFile = saverFile
	}
}

func startMetricServerWorker(
	ctx context.Context,
	lister Lister,
	saver Saver,
	listerFile Lister,
	saverFile Saver,
	restore bool,
	storeInterval int,
) error {
	logger.Log.Debugf("startMetricServerWorker: restore=%v storeInterval=%d", restore, storeInterval)

	if restore {
		logger.Log.Debug("startMetricServerWorker: restoring metrics from file")
		if err := saveMetrics(ctx, listerFile, saver); err != nil {
			logger.Log.Errorw("startMetricServerWorker: restore failed", "error", err)
			return err
		}
		logger.Log.Debug("startMetricServerWorker: restore completed successfully")
	}

	if storeInterval == 0 {
		logger.Log.Debug("startMetricServerWorker: storeInterval=0, waiting for shutdown")
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		logger.Log.Debug("startMetricServerWorker: shutdown triggered, saving metrics")
		if err := saveMetrics(shutdownCtx, listerFile, saver); err != nil {
			logger.Log.Errorw("startMetricServerWorker: error saving metrics during shutdown", "error", err)
			return err
		}
		logger.Log.Debug("startMetricServerWorker: metrics saved successfully on shutdown")
		return nil
	}

	storeTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer storeTicker.Stop()

	logger.Log.Debugf("startMetricServerWorker: starting periodic save every %d seconds", storeInterval)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("startMetricServerWorker: context done, saving metrics before exit")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := saveMetrics(shutdownCtx, lister, saverFile); err != nil {
				logger.Log.Errorw("startMetricServerWorker: error saving metrics on shutdown", "error", err)
				return err
			}
			logger.Log.Debug("startMetricServerWorker: saved metrics successfully on shutdown")
			return nil

		case <-storeTicker.C:
			logger.Log.Debug("startMetricServerWorker: periodic save triggered")
			if err := saveMetrics(ctx, lister, saverFile); err != nil {
				logger.Log.Errorw("startMetricServerWorker: error during periodic save", "error", err)
				return err
			}
			logger.Log.Debug("startMetricServerWorker: periodic save completed successfully")
		}
	}
}

func saveMetrics(
	ctx context.Context,
	lister Lister,
	saver Saver,
) error {
	logger.Log.Debug("saveMetrics: listing metrics")
	metrics, err := lister.List(ctx)
	if err != nil {
		logger.Log.Errorw("saveMetrics: error listing metrics", "error", err)
		return err
	}
	logger.Log.Debugf("saveMetrics: found %d metrics to save", len(metrics))
	for _, metric := range metrics {
		if err := saver.Save(ctx, *metric); err != nil {
			logger.Log.Errorw("saveMetrics: error saving metric", "metricID", metric.ID, "error", err)
			return err
		}
		logger.Log.Debugw("saveMetrics: saved metric successfully", "metricID", metric.ID)
	}
	logger.Log.Debug("saveMetrics: all metrics saved successfully")
	return nil
}
