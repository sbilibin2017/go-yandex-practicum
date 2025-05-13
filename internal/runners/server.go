package runners

import (
	"context"
	"net/http"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

func RunServer(ctx context.Context, server Server) error {
	errCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Log.Info("Shutting down server gracefully...")
		server.Shutdown(ctxShutdown)
		return nil

	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Server failed to start", zap.Error(err))
		}
		return err
	}
}
