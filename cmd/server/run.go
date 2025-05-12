package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	logger.Log.Info("Logger initialized successfully", zap.String("level", flagLogLevel))

	data := make(map[types.MetricID]types.Metrics)

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricMemoryGetByIDRepository := repositories.NewMetricMemoryGetByIDRepository(data)
	metricMemoryListAllRepository := repositories.NewMetricMemoryListAllRepository(data)

	metricUpdateService := services.NewMetricUpdateService(
		metricMemoryGetByIDRepository,
		metricMemorySaveRepository,
	)
	metricGetService := services.NewMetricGetService(
		metricMemoryGetByIDRepository,
	)
	metricListAllService := services.NewMetricListAllService(
		metricMemoryListAllRepository,
	)

	metricUpdateHandler := handlers.MetricUpdatePathHandler(metricUpdateService)
	metricGetHandler := handlers.MetricGetPathHandler(metricGetService)
	metricListAllHandler := handlers.MetricListAllHTMLHandler(metricListAllService)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricUpdateHandler)
	router.Get("/value/{type}/{name}", metricGetHandler)
	router.Get("/", metricListAllHandler)

	server := &http.Server{Addr: flagServerAddress, Handler: router}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Log.Info("Starting server", zap.String("address", flagServerAddress))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Shutting down server gracefully...")
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		return err
	}
	logger.Log.Info("Server stopped gracefully")
	return nil
}
