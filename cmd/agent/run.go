package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

func run() error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	client := resty.New()

	metricFacade := facades.NewMetricFacade(
		client,
		serverAddress,
		header,
		key,
	)

	worker := workers.NewMetricAgentWorker(
		metricFacade,
		pollInterval,
		reportInterval,
		batchSize,
		rateLimit,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := runners.RunWorker(ctx, worker); err != nil {
		return err
	}

	return nil
}
