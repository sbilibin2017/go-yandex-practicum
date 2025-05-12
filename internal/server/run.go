package server

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

func Run(ctx context.Context, server Server) error {
	go func() {
		logger.Log.Info("Starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Shutting down server gracefully...")
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		return err
	}
	logger.Log.Info("Server stopped gracefully")
	return nil
}
