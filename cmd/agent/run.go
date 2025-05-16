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

func run(ctx context.Context) error {
	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	client := resty.New()

	metricFacade := facades.NewMetricFacade(client, flagServerAddress, flagKey)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	worker := workers.NewMetricAgentWorker(
		metricFacade,
		flagPollInterval,
		flagReportInterval,
		flagNumWorkers,
	)

	runners.RunWorker(ctx, worker)

	return nil
}
