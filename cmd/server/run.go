package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/apps"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/server"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	metricsMap := make(map[types.MetricID]types.Metrics)

	router := chi.NewRouter()

	srv := &http.Server{Addr: flagServerAddress}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	apps.ConfigureServerApp(
		ctx,
		metricsMap,
		router,
		srv,
	)

	return server.Run(ctx, srv)

}
