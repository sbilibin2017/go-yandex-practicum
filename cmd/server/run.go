package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func run() error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	var (
		err  error
		file *os.File
	)

	if flagFileStoragePath != "" {
		file, err = os.OpenFile(flagFileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	var db *sqlx.DB
	if flagDatabaseDSN != "" {
		db, err = sqlx.Open("pgx", flagDatabaseDSN)
		if err != nil {
			return err
		}
		if err = goose.SetDialect("postgres"); err != nil {
			logger.Log.Error("Failed to set goose dialect", zap.Error(err))
			return err
		}
		if err = goose.Up(db.DB, "./migrations"); err != nil {
			logger.Log.Error("Failed to apply migrations", zap.Error(err))
			return err
		}
		defer db.Close()
	}

	var (
		metricSaveDBRepo    *repositories.MetricSaveDBRepository
		metricGetByIDDBRepo *repositories.MetricGetByIDDBRepository
		metricListAllDBRepo *repositories.MetricListAllDBRepository
	)
	if db != nil {
		metricSaveDBRepo = repositories.NewMetricSaveDBRepository(db)
		metricGetByIDDBRepo = repositories.NewMetricGetByIDDBRepository(db)
		metricListAllDBRepo = repositories.NewMetricListAllDBRepository(db)
	}

	var (
		metricSaveFileRepo    *repositories.MetricSaveFileRepository
		metricGetByIDFileRepo *repositories.MetricGetByIDFileRepository
		metricListAllFileRepo *repositories.MetricListAllFileRepository
	)
	if file != nil {
		metricSaveFileRepo = repositories.NewMetricSaveFileRepository(file)
		metricGetByIDFileRepo = repositories.NewMetricGetByIDFileRepository(file)
		metricListAllFileRepo = repositories.NewMetricListAllFileRepository(file)
	}

	var (
		memoryStorage           map[types.MetricID]types.Metrics
		metricSaveMemoryRepo    *repositories.MetricSaveMemoryRepository
		metricGetByIDMemoryRepo *repositories.MetricGetByIDMemoryRepository
		metricListAllMemoryRepo *repositories.MetricListAllMemoryRepository
	)
	if db == nil && file == nil {
		memoryStorage = make(map[types.MetricID]types.Metrics)
		metricSaveMemoryRepo = repositories.NewMetricSaveMemoryRepository(memoryStorage)
		metricGetByIDMemoryRepo = repositories.NewMetricGetByIDMemoryRepository(memoryStorage)
		metricListAllMemoryRepo = repositories.NewMetricListAllMemoryRepository(memoryStorage)
	}

	metricSaveContextRepo := repositories.NewMetricSaveContextRepository()
	metricGetByIDContextRepo := repositories.NewMetricGetByIDContextRepository()
	metricListAllContextRepo := repositories.NewMetricListAllContextRepository()

	switch {
	case db != nil:
		metricSaveContextRepo.SetContext(metricSaveDBRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDDBRepo)
		metricListAllContextRepo.SetContext(metricListAllDBRepo)
	case file != nil:
		metricSaveContextRepo.SetContext(metricSaveFileRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDFileRepo)
		metricListAllContextRepo.SetContext(metricListAllFileRepo)
	default:
		metricSaveContextRepo.SetContext(metricSaveMemoryRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDMemoryRepo)
		metricListAllContextRepo.SetContext(metricListAllMemoryRepo)
	}

	metricUpdatesService := services.NewMetricUpdatesService(metricGetByIDContextRepo, metricSaveContextRepo)
	metricGetService := services.NewMetricGetService(metricGetByIDContextRepo)
	metricListAllService := services.NewMetricListAllService(metricListAllContextRepo)

	metricUpdatePathHandler := handlers.NewMetricUpdatePathHandler(metricUpdatesService)
	metricUpdateBodyHandler := handlers.NewMetricUpdateBodyHandler(metricUpdatesService)
	metricUpdatesBodyHandler := handlers.NewMetricUpdatesBodyHandler(metricUpdatesService)
	metricGetPathHandler := handlers.NewMetricGetPathHandler(metricGetService)
	metricGetBodyHandler := handlers.NewMetricGetBodyHandler(metricGetService)
	metricListAllHTMLHandler := handlers.NewMetricListAllHTMLHandler(metricListAllService)
	metricDBPingHandler := handlers.NewDBPingHandler(db)

	metricMiddlewares := []func(next http.Handler) http.Handler{
		middlewares.LoggingMiddleware,
		middlewares.HashMiddleware(flagKey, hashKeyHeader),
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db),
		middlewares.RetryMiddleware,
	}

	metricRouter := chi.NewRouter()
	metricRouter.Use(metricMiddlewares...)
	metricRouter.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	metricRouter.Post("/update/{type}/{name}", metricUpdatePathHandler)
	metricRouter.Post("/update/", metricUpdateBodyHandler)
	metricRouter.Post("/updates/", metricUpdatesBodyHandler)
	metricRouter.Get("/value/{type}/{name}", metricGetPathHandler)
	metricRouter.Post("/value/", metricGetBodyHandler)
	metricRouter.Get("/", metricListAllHTMLHandler)
	metricRouter.Get("/ping", metricDBPingHandler)

	router := chi.NewRouter()
	router.Mount("/", metricRouter)

	server := &http.Server{
		Addr:    flagServerAddress,
		Handler: router,
	}

	var worker func(ctx context.Context) error
	if file != nil {
		worker = workers.NewMetricServerWorker(
			metricListAllContextRepo,
			metricSaveContextRepo,
			metricListAllFileRepo,
			metricSaveFileRepo,
			flagRestore,
			flagStoreInterval,
		)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := runners.RunWorker(ctx, worker); err != nil {
		return err
	}

	if err := runners.RunServer(ctx, server); err != nil {
		return err
	}

	return nil
}
