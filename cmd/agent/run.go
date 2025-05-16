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
	"golang.org/x/sync/semaphore"
)

func run(ctx context.Context) error {
	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	client := resty.New()

	metricFacade := facades.NewMetricFacade(client, flagServerAddress, flagKey)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sem := semaphore.NewWeighted(int64(flagNumWorkers))

	worker := workers.NewMetricAgentWorker(
		metricFacade,
		sem,
		flagPollInterval,
		flagReportInterval,
		flagNumWorkers,
		flagBatchSize,
	)

	runners.RunWorker(ctx, worker)

	return nil
}
