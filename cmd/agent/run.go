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

	client := resty.New()
	metricFacade := facades.NewMetricFacade(*client, flagServerAddress)

	metricCh := make(chan types.MetricUpdatePathRequest, 1000)
	defer close(metricCh)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Log.Info("Starting Metric Agent...")
		workers.StartMetricAgent(
			ctx,
			metricFacade,
			metricCh,
			*pollTicker,
			*reportTicker,
		)
	}()

	<-ctx.Done()

	logger.Log.Info("Shutting down Metric Agent...")

	return nil
}
