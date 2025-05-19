package main

import (
	"context"
	"encoding/json"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/hasher"
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
		json.Marshal,
		hasher.Hash,
		serverAddress,
		key,
		header,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker := workers.NewMetricAgentWorker(
		metricFacade,
		pollInterval,
		reportInterval,
		rateLimit,
		batchSize,
	)

	return runners.RunWorker(ctx, worker)

}
