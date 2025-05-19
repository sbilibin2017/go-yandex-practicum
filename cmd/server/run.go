package main

import (
	"context"
	"os/signal"
	"syscall"

	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/hasher"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/retrier"
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/tx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"

	"go.uber.org/zap"
)

// run инициализирует и запускает HTTP-сервер с конфигурацией и middleware.
// В зависимости от настроек, выбирается хранилище (файл, БД или память).
// Обрабатываются системные сигналы для корректного завершения работы сервера.
func run() error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	var file *os.File
	if fileStoragePath != "" {
		var err error
		file, err = os.OpenFile(fileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	var db *sqlx.DB
	if databaseDSN != "" {
		var err error
		db, err = sqlx.Connect("pgx", databaseDSN)
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
		metricListAllDBRepository = repositories.NewMetricListAllDBRepository(db, tx.GetTxFromContext)
		metricGetByIDDBRepository = repositories.NewMetricGetByIDDBRepository(db, tx.GetTxFromContext)
		metricSaveDBRepository = repositories.NewMetricSaveDBRepository(db, tx.GetTxFromContext)
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
		middlewares.HashMiddleware(key, header, hasher.Hash, hasher.Compare),
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db, tx.WithTx),
		middlewares.DBRetryMiddleware(
			retrier.WithRetry,
			[]time.Duration{
				1 * time.Second,
				3 * time.Second,
				5 * time.Second,
			},
			retrier.IsRetriableDBError,
		),
	)

	router.Post("/update/{type}/{name}/{value}", handlers.NewMetricUpdatePathHandler(metricUpdatesService))
	router.Post("/update/{type}/{name}", handlers.NewMetricUpdatePathHandler(metricUpdatesService))
	router.Post("/update/", handlers.NewMetricUpdateBodyHandler(metricUpdatesService))
	router.Post("/updates/", handlers.NewMetricUpdatesBodyHandler(metricUpdatesService))
	router.Get("/value/{type}/{name}", handlers.NewMetricGetPathHandler(metricGetService))
	router.Post("/value/", handlers.NewMetricGetBodyHandler(metricGetService))
	router.Get("/", handlers.NewMetricListAllHTMLHandler(metricListAllService))
	router.Get("/ping", handlers.NewDBPingHandler(db))

	server := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if metricListAllFileRepository != nil && metricSaveFileRepository != nil {
		worker := workers.NewMetricServerWorker(
			ctx,
			metricListAllContextRepository,
			metricSaveContextRepository,
			metricListAllFileRepository,
			metricSaveFileRepository,
			restore,
			storeInterval,
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
