package main

import (
	"context"
	"net/http"
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
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories/db"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories/file"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories/memory"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories/strategies"

	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
	"go.uber.org/zap"
)

var (
	conn *sqlx.DB

	metricSaveDBRepo    *db.MetricSaveDBRepository
	metricGetByIDDBRepo *db.MetricGetByIDDBRepository
	metricListAllDBRepo *db.MetricListAllDBRepository

	metricSaveFileRepo    *file.MetricSaveFileRepository
	metricGetByIDFileRepo *file.MetricGetByIDFileRepository
	metricListAllFileRepo *file.MetricListAllFileRepository

	metricSaveMemoryRepo    *memory.MetricSaveMemoryRepository
	metricGetByIDMemoryRepo *memory.MetricGetByIDMemoryRepository
	metricListAllMemoryRepo *memory.MetricListAllMemoryRepository

	metricSaveContextRepo    *strategies.MetricSaveContextRepository
	metricGetByIDContextRepo *strategies.MetricGetByIDContextRepository
	metricListAllContextRepo *strategies.MetricListAllContextRepository

	metricUpdatesService *services.MetricUpdatesService
	metricGetService     *services.MetricGetService
	metricListAllService *services.MetricListAllService

	metricUpdatePathHandler  http.HandlerFunc
	metricUpdateBodyHandler  http.HandlerFunc
	metricUpdatesBodyHandler http.HandlerFunc
	metricGetPathHandler     http.HandlerFunc
	metricGetBodyHandler     http.HandlerFunc
	metricListAllHTMLHandler http.HandlerFunc

	dbPingHandler http.HandlerFunc

	metricRouter *chi.Mux

	startMetricServerWorkerFunc = workers.StartMetricServerWorker

	startHTTPServerFunc = func(server *http.Server, errChan chan<- error) {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}
)

func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	var err error

	if flagDatabaseDSN != "" {
		conn, err = sqlx.Open("pgx", flagDatabaseDSN)
		if err != nil {
			return err
		}
		if err = goose.SetDialect("postgres"); err != nil {
			logger.Log.Error("Failed to set goose dialect", zap.Error(err))
			return err
		}
		if err = goose.Up(conn.DB, "./migrations"); err != nil {
			logger.Log.Error("Failed to apply migrations", zap.Error(err))
			return err
		}
		defer conn.Close()
	}

	if flagDatabaseDSN != "" {
		metricSaveDBRepo = db.NewMetricSaveDBRepository(conn, middlewares.GetExecutor)
		metricGetByIDDBRepo = db.NewMetricGetByIDDBRepository(conn, middlewares.GetExecutor)
		metricListAllDBRepo = db.NewMetricListAllDBRepository(conn, middlewares.GetExecutor)
	}

	if flagDatabaseDSN == "" && flagFileStoragePath != "" {
		metricSaveFileRepo = file.NewMetricSaveFileRepository(flagFileStoragePath)
		metricGetByIDFileRepo = file.NewMetricGetByIDFileRepository(flagFileStoragePath)
		metricListAllFileRepo = file.NewMetricListAllFileRepository(flagFileStoragePath)
	}

	if flagDatabaseDSN == "" && flagFileStoragePath == "" {
		metricSaveMemoryRepo = memory.NewMetricSaveMemoryRepository()
		metricGetByIDMemoryRepo = memory.NewMetricGetByIDMemoryRepository()
		metricListAllMemoryRepo = memory.NewMetricListAllMemoryRepository()
	}

	metricSaveContextRepo = strategies.NewMetricSaveContextRepository()
	metricGetByIDContextRepo = strategies.NewMetricGetByIDContextRepository()
	metricListAllContextRepo = strategies.NewMetricListAllContextRepository()

	switch {
	case flagDatabaseDSN != "":
		metricSaveContextRepo.SetContext(metricSaveDBRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDDBRepo)
		metricListAllContextRepo.SetContext(metricListAllDBRepo)
	case flagFileStoragePath != "":
		metricSaveContextRepo.SetContext(metricSaveFileRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDFileRepo)
		metricListAllContextRepo.SetContext(metricListAllFileRepo)
	default:
		metricSaveContextRepo.SetContext(metricSaveMemoryRepo)
		metricGetByIDContextRepo.SetContext(metricGetByIDMemoryRepo)
		metricListAllContextRepo.SetContext(metricListAllMemoryRepo)
	}

	metricUpdatesService = services.NewMetricUpdatesService(metricGetByIDContextRepo, metricSaveContextRepo)
	metricGetService = services.NewMetricGetService(metricGetByIDContextRepo)
	metricListAllService = services.NewMetricListAllService(metricListAllContextRepo)

	metricUpdatePathHandler = handlers.NewMetricUpdatePathHandler(metricUpdatesService)
	metricUpdateBodyHandler = handlers.NewMetricUpdateBodyHandler(metricUpdatesService)
	metricUpdatesBodyHandler = handlers.NewMetricUpdatesBodyHandler(metricUpdatesService)
	metricGetPathHandler = handlers.NewMetricGetPathHandler(metricGetService)
	metricGetBodyHandler = handlers.NewMetricGetBodyHandler(metricGetService)
	metricListAllHTMLHandler = handlers.NewMetricListAllHTMLHandler(metricListAllService)

	dbPingHandler = newDBPingHandler(conn)

	metricMiddlewares := []func(http.Handler) http.Handler{
		middlewares.TrustedSubnetMiddleware(flagTrustedSubnet),
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
		middlewares.CryptoMiddleware(flagCryptoKey),
		middlewares.HashMiddleware(flagKey, hashKeyHeader),
		middlewares.TxMiddleware(conn),
		middlewares.RetryMiddleware,
	}

	metricRouter = chi.NewRouter()
	metricRouter.Use(metricMiddlewares...)

	metricRouter.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	metricRouter.Post("/update/{type}/{name}", metricUpdatePathHandler)
	metricRouter.Post("/update/", metricUpdateBodyHandler)
	metricRouter.Post("/updates/", metricUpdatesBodyHandler)
	metricRouter.Get("/value/{type}/{name}", metricGetPathHandler)
	metricRouter.Post("/value/", metricGetBodyHandler)
	metricRouter.Get("/", metricListAllHTMLHandler)
	metricRouter.Get("/ping", dbPingHandler)

	router := chi.NewRouter()
	router.Mount("/", metricRouter)

	server := &http.Server{
		Addr:    flagServerAddress,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	errChan := make(chan error, 1)

	if flagFileStoragePath != "" {
		go func() {
			if err := startMetricServerWorkerFunc(
				ctx,
				metricListAllContextRepo,
				metricSaveContextRepo,
				metricListAllFileRepo,
				metricSaveFileRepo,
				flagRestore,
				flagStoreInterval,
			); err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		logger.Log.Info("Starting server", zap.String("addr", flagServerAddress))
		startHTTPServerFunc(server, errChan)
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("Received shutdown signal")
	case err := <-errChan:
		if err != nil {
			logger.Log.Error("Background worker error", zap.Error(err))
		}
	}

	logger.Log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		return err
	}

	logger.Log.Info("Server exited gracefully")
	return nil
}

func newDBPingHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
