package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, opts *options) error {
	if err := logger.Initialize(opts.LogLevel); err != nil {
		return err
	}

	var file *os.File
	if opts.FileStoragePath != "" {
		var err error
		file, err = os.OpenFile(opts.FileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	var db *sqlx.DB
	if opts.DatabaseDSN != "" {
		var err error
		db, err = sqlx.Connect("pgx", opts.DatabaseDSN)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := goose.SetDialect("postgres"); err != nil {
			logger.Log.Error("Failed to set goose dialect", zap.Error(err))
			return err
		}
		if err := goose.Up(db.DB, "./migrations"); err != nil {
			logger.Log.Error("Failed to apply migrations", zap.Error(err))
			return err
		}
	}

	var storeTicker *time.Ticker
	if opts.StoreInterval > 0 {
		storeTicker = time.NewTicker(time.Duration(opts.StoreInterval) * time.Second)
		defer storeTicker.Stop()
	}

	var (
		metricListAllFileRepository *repositories.MetricListAllFileRepository
		metricGetByIDFileRepository *repositories.MetricGetByIDFileRepository
		metricSaveFileRepository    *repositories.MetricSaveFileRepository
	)
	if file != nil {
		metricListAllFileRepository = repositories.NewMetricListAllFileRepository(file)
		metricGetByIDFileRepository = repositories.NewMetricGetByIDFileRepository(file)
		metricSaveFileRepository = repositories.NewMetricSaveFileRepository(file)
	}

	var (
		metricListAllDBRepository *repositories.MetricListAllDBRepository
		metricGetByIDDBRepository *repositories.MetricGetByIDDBRepository
		metricSaveDBRepository    *repositories.MetricSaveDBRepository
	)
	if db != nil {
		metricListAllDBRepository = repositories.NewMetricListAllDBRepository(db, middlewares.GetTx)
		metricGetByIDDBRepository = repositories.NewMetricGetByIDDBRepository(db, middlewares.GetTx)
		metricSaveDBRepository = repositories.NewMetricSaveDBRepository(db, middlewares.GetTx)
	}

	var (
		metricListAllMemoryRepository *repositories.MetricListAllMemoryRepository
		metricGetByIDMemoryRepository *repositories.MetricGetByIDMemoryRepository
		metricSaveMemoryRepository    *repositories.MetricSaveMemoryRepository
	)
	if file == nil && db == nil {
		metricsMap := make(map[types.MetricID]types.Metrics)
		metricListAllMemoryRepository = repositories.NewMetricListAllMemoryRepository(metricsMap)
		metricGetByIDMemoryRepository = repositories.NewMetricGetByIDMemoryRepository(metricsMap)
		metricSaveMemoryRepository = repositories.NewMetricSaveMemoryRepository(metricsMap)
	}

	metricListAllContextRepository := repositories.NewMetricListAllContextRepository()
	metricGetByIDContextRepository := repositories.NewMetricGetByIDContextRepository()
	metricSaveContextRepository := repositories.NewMetricSaveContextRepository()

	switch {
	case db != nil:
		metricListAllContextRepository.SetContext(metricListAllDBRepository)
		metricGetByIDContextRepository.SetContext(metricGetByIDDBRepository)
		metricSaveContextRepository.SetContext(metricSaveDBRepository)
	case file != nil:
		metricListAllContextRepository.SetContext(metricListAllFileRepository)
		metricGetByIDContextRepository.SetContext(metricGetByIDFileRepository)
		metricSaveContextRepository.SetContext(metricSaveFileRepository)
	default:
		metricListAllContextRepository.SetContext(metricListAllMemoryRepository)
		metricGetByIDContextRepository.SetContext(metricGetByIDMemoryRepository)
		metricSaveContextRepository.SetContext(metricSaveMemoryRepository)
	}

	metricUpdateService := services.NewMetricUpdateService(metricGetByIDContextRepository, metricSaveContextRepository)
	metricGetService := services.NewMetricGetService(metricGetByIDContextRepository)
	metricListAllService := services.NewMetricListAllService(metricListAllContextRepository)

	router := chi.NewRouter()
	router.Use(
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db),
	)
	router.Post("/update/{type}/{name}/{value}", handlers.NewMetricUpdatePathHandler(metricUpdateService))
	router.Post("/update/", handlers.NewMetricUpdateBodyHandler(metricUpdateService))
	router.Get("/value/{type}/{name}", handlers.NewMetricGetPathHandler(metricGetService))
	router.Post("/value/", handlers.NewMetricGetBodyHandler(metricGetService))
	router.Get("/", handlers.NewMetricListAllHTMLHandler(metricListAllService))
	router.Get("/ping", handlers.NewDBPingHandler(db))

	server := &http.Server{
		Addr:    opts.ServerAddress,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grp, ctx := errgroup.WithContext(ctx)

	if metricListAllFileRepository != nil && metricSaveFileRepository != nil {
		grp.Go(func() error {
			return workers.StartMetricServerWorker(
				ctx,
				metricListAllContextRepository,
				metricSaveContextRepository,
				metricListAllFileRepository,
				metricSaveFileRepository,
				storeTicker,
				opts.Restore,
			)
		})
	}

	grp.Go(func() error {
		logger.Log.Info("Starting HTTP server", zap.String("addr", server.Addr))

		errCh := make(chan error, 1)
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Log.Error("Server ListenAndServe error", zap.Error(err))
				errCh <- err
			} else {
				logger.Log.Info("Server stopped listening", zap.Error(err))
			}
			close(errCh)
		}()

		select {
		case <-ctx.Done():
			logger.Log.Info("Context done, shutting down server")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Log.Error("Error during server shutdown", zap.Error(err))
				return err
			}
			logger.Log.Info("Server shutdown complete")
			return nil

		case err := <-errCh:
			if err != nil {
				logger.Log.Error("Server error received from errCh", zap.Error(err))
			} else {
				logger.Log.Info("Server exited without error")
			}
			return err
		}
	})

	err := grp.Wait()
	return err
}
