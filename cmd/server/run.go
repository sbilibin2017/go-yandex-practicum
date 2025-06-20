package main

import (
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/routers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/validators"
)

var (
	db *sqlx.DB

	metricSaveDBRepository    *repositories.MetricSaveDBRepository
	metricGetByIDDBRepository *repositories.MetricGetByIDDBRepository
	metricListAllDBRepository *repositories.MetricListAllDBRepository

	metricSaveFileRepository    *repositories.MetricSaveFileRepository
	metricGetByIDFileRepository *repositories.MetricGetByIDFileRepository
	metricListAllFileRepository *repositories.MetricListAllFileRepository

	metricSaveMemoryRepository    *repositories.MetricSaveMemoryRepository
	metricGetByIDMemoryRepository *repositories.MetricGetByIDMemoryRepository
	metricListAllMemoryRepository *repositories.MetricListAllMemoryRepository

	metricSaveContextRepo    *repositories.MetricSaveContextRepository
	metricGetByIDContextRepo *repositories.MetricGetByIDContextRepository
	metricListAllContextRepo *repositories.MetricListAllContextRepository

	metricUpdatesService *services.MetricUpdatesService
	metricGetService     *services.MetricGetService
	metricListAllService *services.MetricListAllService

	valMetric       func(*types.Metrics) error
	valMetricID     func(types.MetricID) error
	valMetricAttr   func(mType string, mName string, mValue string) error
	valMetricIDAttr func(mType string, mID string) error

	metricUpdatePathHandler  *handlers.MetricUpdatePathHandler
	metricUpdateBodyHandler  *handlers.MetricUpdateBodyHandler
	metricUpdatesBodyHandler *handlers.MetricUpdatesBodyHandler
	metricGetPathHandler     *handlers.MetricGetPathHandler
	metricGetBodyHandler     *handlers.MetricGetBodyHandler
	metricListAllHTMLHandler *handlers.MetricListAllHTMLHandler

	router *chi.Mux

	srv *http.Server
)

func run() error {
	err := logger.Initialize(logLevel)
	if err != nil {
		return err
	}

	// Initialize DB connection if DSN is provided
	if flagDatabaseDSN != "" {
		db, err = sqlx.Open("pgx", flagDatabaseDSN)
		if err != nil {
			return err
		}

		if err = goose.SetDialect("postgres"); err != nil {
			return err
		}

		if err = goose.Up(db.DB, migrationsDir); err != nil {
			return err
		}
	}

	// Initialize repositories depending on config
	if flagDatabaseDSN != "" {
		metricSaveDBRepository = repositories.NewMetricSaveDBRepository(db, middlewares.GetTx)
		metricGetByIDDBRepository = repositories.NewMetricGetByIDDBRepository(db, middlewares.GetTx)
		metricListAllDBRepository = repositories.NewMetricListAllDBRepository(db, middlewares.GetTx)
	}

	if flagDatabaseDSN == "" && flagFileStoragePath != "" {
		metricSaveFileRepository = repositories.NewMetricSaveFileRepository(flagFileStoragePath)
		metricGetByIDFileRepository = repositories.NewMetricGetByIDFileRepository(flagFileStoragePath)
		metricListAllFileRepository = repositories.NewMetricListAllFileRepository(flagFileStoragePath)
	}

	if flagDatabaseDSN == "" && flagFileStoragePath == "" {
		metricSaveMemoryRepository = repositories.NewMetricSaveMemoryRepository()
		metricGetByIDMemoryRepository = repositories.NewMetricGetByIDMemoryRepository()
		metricListAllMemoryRepository = repositories.NewMetricListAllMemoryRepository()
	}

	// Initialize context repositories
	metricSaveContextRepo = repositories.NewMetricSaveContextRepository()
	metricGetByIDContextRepo = repositories.NewMetricGetByIDContextRepository()
	metricListAllContextRepo = repositories.NewMetricListAllContextRepository()

	// Wire context repositories to concrete implementations
	switch {
	case flagDatabaseDSN != "":
		metricSaveContextRepo.SetContext(metricSaveDBRepository)
		metricGetByIDContextRepo.SetContext(metricGetByIDDBRepository)
		metricListAllContextRepo.SetContext(metricListAllDBRepository)
	case flagFileStoragePath != "":
		metricSaveContextRepo.SetContext(metricSaveFileRepository)
		metricGetByIDContextRepo.SetContext(metricGetByIDFileRepository)
		metricListAllContextRepo.SetContext(metricListAllFileRepository)
	default:
		metricSaveContextRepo.SetContext(metricSaveMemoryRepository)
		metricGetByIDContextRepo.SetContext(metricGetByIDMemoryRepository)
		metricListAllContextRepo.SetContext(metricListAllMemoryRepository)
	}

	// Initialize validators
	valMetric = validators.ValidateMetric
	valMetricID = validators.ValidateMetricID
	valMetricAttr = validators.ValidateMetricAttributes
	valMetricIDAttr = validators.ValidateMetricIDAttributes

	// Initialize services
	metricUpdatesService = services.NewMetricUpdatesService(metricGetByIDContextRepo, metricSaveContextRepo)
	metricGetService = services.NewMetricGetService(metricGetByIDContextRepo)
	metricListAllService = services.NewMetricListAllService(metricListAllContextRepo)

	// Initialize HTTP handlers with the services and validators
	metricUpdatePathHandler = handlers.NewMetricUpdatePathHandler(
		metricUpdatesService,
		valMetricAttr,
	)
	metricUpdateBodyHandler = handlers.NewMetricUpdateBodyHandler(
		metricUpdatesService,
		valMetric,
	)
	metricUpdatesBodyHandler = handlers.NewMetricUpdatesBodyHandler(
		metricUpdatesService,
		valMetric,
	)
	metricGetPathHandler = handlers.NewMetricGetPathHandler(
		metricGetService,
		valMetricIDAttr,
	)
	metricGetBodyHandler = handlers.NewMetricGetBodyHandler(
		metricGetService,
		valMetricID,
	)
	metricListAllHTMLHandler = handlers.NewMetricListAllHTMLHandler(
		metricListAllService,
	)

	// Middleware chain
	metricMiddlewares := []func(http.Handler) http.Handler{
		middlewares.TrustedSubnetMiddleware(flagTrustedSubnet),
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
		middlewares.CryptoMiddleware(flagCryptoKey),
		middlewares.HashMiddleware(flagKey, flagHashKeyHeader),
		middlewares.TxMiddleware(db),
		middlewares.RetryMiddleware,
	}

	// Setup router
	router = routers.NewMetricRouter(
		metricUpdatePathHandler,
		metricUpdateBodyHandler,
		metricUpdatesBodyHandler,
		metricGetPathHandler,
		metricGetBodyHandler,
		metricListAllHTMLHandler,
		metricMiddlewares...,
	)

	// Start server
	srv = &http.Server{
		Addr:    flagServerAddress,
		Handler: router,
	}

	return nil
}
