package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
	"go.uber.org/zap"
)

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	metricsMap := make(map[types.MetricID]types.Metrics)

	var file *os.File
	if flagFileStoragePath != "" {
		file, err = os.OpenFile(flagFileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
	}

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(metricsMap)
	metricMemoryGetByIDRepository := repositories.NewMetricMemoryGetByIDRepository(metricsMap)
	metricMemoryListAllRepository := repositories.NewMetricMemoryListAllRepository(metricsMap)
	metricFileListAllRepository := repositories.NewMetricFileListAllRepository(file)
	metricFileSaveRepository := repositories.NewMetricFileSaveRepository(file)

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

	metricUpdatePathHandler := handlers.NewMetricUpdatePathHandler(metricUpdateService)
	metricUpdateBodyHandler := handlers.NewMetricUpdateBodyHandler(metricUpdateService)
	metricGetPathHandler := handlers.NewMetricGetPathHandler(metricGetService)
	metricGetBodyHandler := handlers.NewMetricGetBodyHandler(metricGetService)
	metricListAllHandler := handlers.NewMetricListAllHTMLHandler(metricListAllService)

	metricRouter := chi.NewRouter()
	metricRouter.Use(
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
	)
	metricRouter.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	metricRouter.Post("/update/", metricUpdateBodyHandler)
	metricRouter.Get("/value/{type}/{name}", metricGetPathHandler)
	metricRouter.Post("/value/", metricGetBodyHandler)
	metricRouter.Get("/", metricListAllHandler)

	router := chi.NewRouter()
	router.Mount("/", metricRouter)

	server := &http.Server{Addr: flagServerAddress, Handler: router}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var storeTicker *time.Ticker
	if flagStoreInterval != 0 {
		storeTicker = time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
	}

	go workers.StartMetricServerWorker(
		ctx,
		metricMemoryListAllRepository,
		metricMemorySaveRepository,
		metricFileListAllRepository,
		metricFileSaveRepository,
		storeTicker,
		flagRestore,
	)

	errCh := make(chan error, 1)

	go func() {
		logger.Log.Info("Starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Log.Info("Shutting down server gracefully...")
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Log.Error("Server shutdown error", zap.Error(err))
		}

	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Server failed to start", zap.Error(err))
			return err
		}
	}

	if storeTicker != nil {
		storeTicker.Stop()
	}
	if file != nil {
		file.Close()
		logger.Log.Info("Storage file closed")
	}

	return nil
}
