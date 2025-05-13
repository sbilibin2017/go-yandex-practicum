package runners

import (
	"context"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

type Worker interface {
	Start(ctx context.Context) error
}

func RunWorker(ctx context.Context, worker Worker) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- worker.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutting down Metric Agent...")
		return nil
	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Worker exited with error: ", zap.Error(err))
			return err
		}
		logger.Log.Info("Worker exited cleanly")
		return nil
	}
}
