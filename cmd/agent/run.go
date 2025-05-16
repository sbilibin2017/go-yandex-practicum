package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

func run(ctx context.Context, opts *options) error {
	if err := logger.Initialize(opts.LogLevel); err != nil {
		return err
	}

	client := resty.New()
	metricFacade := facades.NewMetricFacade(client, opts.ServerAddress, opts.Key)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go workers.StartMetricAgentWorker(
		ctx,
		metricFacade,
		opts.PollInterval,
		opts.ReportInterval,
		opts.NumWorkers,
	)

	return nil
}
