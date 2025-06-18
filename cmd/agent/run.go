package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

// Declare startMetricAgentFunc with exact signature as StartMetricAgent
var startMetricAgentFunc = func(
	ctx context.Context,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	pollInterval int,
	reportInterval int,
	batchSize int,
	rateLimit int,
) error {
	return workers.StartMetricAgent(
		ctx,
		serverAddress,
		header,
		key,
		cryptoKeyPath,
		pollInterval,
		reportInterval,
		batchSize,
		rateLimit,
	)
}

func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		errCh <- startMetricAgentFunc(
			ctx,
			flagServerAddress,
			hashKeyHeader,
			flagKey,
			flagCryptoKey,
			flagPollInterval,
			flagReportInterval,
			batchSize,
			flagRateLimit,
		)
	}()

	select {
	case <-ctx.Done():
		// Graceful shutdown
		return nil
	case err := <-errCh:
		return err
	}
}
