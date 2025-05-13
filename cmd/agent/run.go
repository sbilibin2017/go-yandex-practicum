package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
	"go.uber.org/zap"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	metricsCh := make(chan types.Metrics, 1000)
	defer close(metricsCh)

	pollTicker := time.NewTicker(time.Duration(flagPollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(flagReportInterval) * time.Second)
	defer reportTicker.Stop()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := resty.New()

	metricFacade := facades.NewMetricFacade(client, flagServerAddress, flagServerUpdateEndpoint)

	errCh := make(chan error, 1)

	go func() {
		errCh <- workers.StartMetricAgentWorker(
			ctx,
			metricFacade,
			metricsCh,
			pollTicker,
			reportTicker,
		)
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
