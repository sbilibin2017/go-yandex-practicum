package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/apps"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	pollTicker := time.NewTicker(time.Duration(flagPollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(flagReportInterval) * time.Second)
	defer reportTicker.Stop()

	metricsCh := make(chan types.MetricUpdatePathRequest, 1000)
	defer close(metricsCh)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := resty.New()

	metricAgent := apps.ConfigureAgentApp(
		*client,
		flagServerAddress,
		metricsCh,
		pollTicker,
		reportTicker,
	)

	return runners.RunWorker(ctx, metricAgent)
}
