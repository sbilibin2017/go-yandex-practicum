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
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
	"go.uber.org/zap"
)

func run(ctx context.Context) error {
	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	var file *os.File
	if flagFileStoragePath != "" {
		var err error
		file, err = os.OpenFile(flagFileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	var db *sqlx.DB
	if flagDatabaseDSN != "" {
		var err error
		db, err = sqlx.Connect("pgx", flagDatabaseDSN)
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
	if flagStoreInterval > 0 {
		storeTicker = time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
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

	metricUpdatesService := services.NewMetricUpdatesService(metricGetByIDContextRepository, metricSaveContextRepository)
	metricGetService := services.NewMetricGetService(metricGetByIDContextRepository)
	metricListAllService := services.NewMetricListAllService(metricListAllContextRepository)

	router := chi.NewRouter()

	router.Use(
		middlewares.LoggingMiddleware,
		middlewares.HashMiddleware(flagKey, flagHeader),
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db, middlewares.SetTx),
		middlewares.DBRetryMiddleware,
	)

	router.Post("/update/{type}/{name}/{value}", handlers.NewMetricUpdatePathHandler(metricUpdatesService))
	router.Post("/update/", handlers.NewMetricUpdateBodyHandler(metricUpdatesService))
	router.Post("/updates/", handlers.NewMetricUpdatesBodyHandler(metricUpdatesService))
	router.Get("/value/{type}/{name}", handlers.NewMetricGetPathHandler(metricGetService))
	router.Post("/value/", handlers.NewMetricGetBodyHandler(metricGetService))
	router.Get("/", handlers.NewMetricListAllHTMLHandler(metricListAllService))
	router.Get("/ping", handlers.NewDBPingHandler(db))

	server := &http.Server{
		Addr:    flagServerAddress,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if metricListAllFileRepository != nil && metricSaveFileRepository != nil {
		worker := workers.NewMetricServerWorker(
			ctx,
			metricListAllContextRepository,
			metricSaveContextRepository,
			metricListAllFileRepository,
			metricSaveFileRepository,
			flagRestore,
			flagStoreInterval,
		)

		err := runners.RunWorker(ctx, worker)
		if err != nil {
			return err
		}
	}

	err := runners.RunServer(ctx, server)
	if err != nil {
		return err
	}

	return nil
}
