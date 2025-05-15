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
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, opts *options) error {
	if err := logger.Initialize(opts.LogLevel); err != nil {
		return err
	}

	metricsCh := make(chan types.Metrics, 1000)
	defer close(metricsCh)

	pollTicker := time.NewTicker(time.Duration(opts.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(opts.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	client := resty.New()
	metricFacade := facades.NewMetricFacade(client, opts.ServerAddress)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		err := workers.StartMetricAgentWorker(ctx, metricFacade, metricsCh, pollTicker, reportTicker)
		if err != nil {
			logger.Log.Error("Metric agent worker stopped with error", zap.Error(err))
		} else {
			logger.Log.Info("Metric agent worker stopped gracefully")
		}
		return err
	})

	err := grp.Wait()

	if err != nil {
		logger.Log.Error("Worker returned error", zap.Error(err))
	} else {
		logger.Log.Info("Shutdown gracefully")
	}

	return err
}
