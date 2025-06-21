package apps

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/sbilibin2017/go-yandex-practicum/internal/contexts"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

// ServerAppConfig holds the configuration parameters for the server application.
type ServerAppConfig struct {
	ServerAddress   string // address and port to run the server on
	DatabaseDSN     string // data source name for database connection
	StoreInterval   int    // interval in seconds to store data
	FileStoragePath string // path to store files on disk
	Restore         bool   // whether to restore data from backup on startup
	Key             string // key used for hashing or encryption
	CryptoKey       string // path to private key file for encryption
	ConfigPath      string // path to external config file
	TrustedSubnet   string // trusted subnet in CIDR notation
	HashHeader      string // HTTP header for SHA256 hash
	LogLevel        string // logging level (e.g., debug, info)
	MigrationsDir   string // directory containing DB migration files
}

// ServerAppOpt defines a functional option for configuring ServerAppConfig.
type ServerAppOpt func(*ServerAppConfig)

// NewServerAppConfig creates a new ServerAppConfig with the provided options.
func NewServerAppConfig(opts ...ServerAppOpt) *ServerAppConfig {
	cfg := &ServerAppConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// WithServerAddress sets the server address (host:port) for the application.
func WithServerAddress(addr string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.ServerAddress = addr
	}
}

// WithServerDatabaseDSN sets the Data Source Name (DSN) for the database connection.
func WithServerDatabaseDSN(dsn string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.DatabaseDSN = dsn
	}
}

// WithServerStoreInterval sets the interval in seconds to store metrics data.
func WithServerStoreInterval(interval int) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.StoreInterval = interval
	}
}

// WithServerFileStoragePath sets the file path for storing metrics on disk.
func WithServerFileStoragePath(path string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.FileStoragePath = path
	}
}

// WithServerRestore enables or disables restoring metrics from file on startup.
func WithServerRestore(restore bool) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.Restore = restore
	}
}

// WithServerKey sets the application key used for hashing or encryption.
func WithServerKey(key string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.Key = key
	}
}

// WithServerCryptoKey sets the path to the private key used for encryption.
func WithServerCryptoKey(cryptoKey string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.CryptoKey = cryptoKey
	}
}

// WithServerConfigPath sets the path to the external configuration file.
func WithServerConfigPath(configPath string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.ConfigPath = configPath
	}
}

// WithServerTrustedSubnet sets the trusted subnet in CIDR notation.
func WithServerTrustedSubnet(subnet string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.TrustedSubnet = subnet
	}
}

// WithServerHashHeader sets the name of the HTTP header used to pass the SHA256 hash.
func WithServerHashHeader(hashHeader string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.HashHeader = hashHeader
	}
}

// WithServerLogLevel sets the logging level (e.g., debug, info, warn, error).
func WithServerLogLevel(logLevel string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.LogLevel = logLevel
	}
}

// WithServerMigrationsDir sets the directory path containing database migration files.
func WithServerMigrationsDir(migrationsDir string) ServerAppOpt {
	return func(c *ServerAppConfig) {
		c.MigrationsDir = migrationsDir
	}
}

// ServerApp represents the main application server, including configuration,
// services, repositories, handlers, and HTTP server.
type ServerApp struct {
	Config *ServerAppConfig

	DB *sqlx.DB

	MetricDBSaveRepository *repositories.MetricDBSaveRepository
	MetricDBGetRepository  *repositories.MetricDBGetRepository
	MetricDBListRepository *repositories.MetricDBListRepository

	MetricFileSaveRepository *repositories.MetricFileSaveRepository
	MetricFileGetRepository  *repositories.MetricFileGetRepository
	MetricFileListRepository *repositories.MetricFileListRepository

	MetricMemorySaveRepository *repositories.MetricMemorySaveRepository
	MetricMemoryGetRepository  *repositories.MetricMemoryGetRepository
	MetricMemoryListRepository *repositories.MetricMemoryListRepository

	MetricContextSaveRepository *repositories.MetricContextSaveRepository
	MetricContextGetRepository  *repositories.MetricContextGetRepository
	MetricContextListRepository *repositories.MetricContextListRepository

	MetricUpdatesService *services.MetricUpdatesService
	MetricGetService     *services.MetricGetService
	MetricListService    *services.MetricListService

	MetricUpdatePathHandler  *handlers.MetricUpdatePathHandler
	MetricUpdateBodyHandler  *handlers.MetricUpdateBodyHandler
	MetricUpdatesBodyHandler *handlers.MetricUpdatesBodyHandler
	MetricGetPathHandler     *handlers.MetricGetPathHandler
	MetricGetBodyHandler     *handlers.MetricGetBodyHandler
	MetricListHTMLHandler    *handlers.MetricListHTMLHandler

	PingHandlerHandler *handlers.PingDBHandler

	Workers []func(ctx context.Context) error
	Router  *chi.Mux
	Srv     *http.Server
}

// NewServerApp creates and initializes a new instance of ServerApp.
func NewServerApp(opts ...ServerAppOpt) (*ServerApp, error) {
	cfg := NewServerAppConfig(opts...)

	err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		logger.Log.Error("Failed to initialize logger: " + err.Error())
		return nil, err
	}

	var app ServerApp
	app.Config = cfg

	if cfg.DatabaseDSN != "" {
		app.DB, err = sqlx.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		if cfg.MigrationsDir != "" {
			if err := goose.SetDialect("postgres"); err != nil {
				return nil, err
			}
			if err := goose.Up(app.DB.DB, cfg.MigrationsDir); err != nil {
				return nil, err
			}
		}

		logger.Log.Info("Initializing DB repositories")
		app.MetricDBSaveRepository = repositories.NewMetricDBSaveRepository(
			repositories.WithMetricDBSaveRepositoryDB(app.DB),
			repositories.WithMetricDBSaveRepositoryTxGetter(contexts.GetTxFromContext),
		)
		app.MetricDBGetRepository = repositories.NewMetricDBGetRepository(
			repositories.WithMetricDBGetRepositoryDB(app.DB),
			repositories.WithMetricDBGetRepositoryTxGetter(contexts.GetTxFromContext),
		)
		app.MetricDBListRepository = repositories.NewMetricDBListRepository(
			repositories.WithMetricDBListRepositoryDB(app.DB),
			repositories.WithMetricDBListRepositoryTxGetter(contexts.GetTxFromContext),
		)
	}

	if cfg.DatabaseDSN == "" && cfg.FileStoragePath != "" {
		logger.Log.Infof("Using file storage at path: %s", cfg.FileStoragePath)
		if err := os.MkdirAll(filepath.Dir(cfg.FileStoragePath), 0755); err != nil {
			return nil, err
		}
		app.MetricFileSaveRepository = repositories.NewMetricFileSaveRepository(
			repositories.WithMetricFileSaveRepositoryPath(cfg.FileStoragePath),
		)
		app.MetricFileGetRepository = repositories.NewMetricFileGetRepository(
			repositories.WithMetricFileGetRepositoryPath(cfg.FileStoragePath),
		)
		app.MetricFileListRepository = repositories.NewMetricFileListRepository(
			repositories.WithMetricFileListRepositoryPath(cfg.FileStoragePath),
		)
	}

	if cfg.DatabaseDSN == "" && cfg.FileStoragePath == "" {
		logger.Log.Info("Using in-memory storage")
		app.MetricMemorySaveRepository = repositories.NewMetricMemorySaveRepository()
		app.MetricMemoryGetRepository = repositories.NewMetricMemoryGetRepository()
		app.MetricMemoryListRepository = repositories.NewMetricMemoryListRepository()
	}

	app.MetricContextSaveRepository = repositories.NewMetricContextSaveRepository()
	app.MetricContextGetRepository = repositories.NewMetricContextGetRepository()
	app.MetricContextListRepository = repositories.NewMetricContextListRepository()

	if app.MetricDBSaveRepository != nil {
		app.MetricContextSaveRepository.SetContext(app.MetricDBSaveRepository)
		app.MetricContextGetRepository.SetContext(app.MetricDBGetRepository)
		app.MetricContextListRepository.SetContext(app.MetricDBListRepository)
	} else if app.MetricFileSaveRepository != nil {
		app.MetricContextSaveRepository.SetContext(app.MetricFileSaveRepository)
		app.MetricContextGetRepository.SetContext(app.MetricFileGetRepository)
		app.MetricContextListRepository.SetContext(app.MetricFileListRepository)
	} else {
		app.MetricContextSaveRepository.SetContext(app.MetricMemorySaveRepository)
		app.MetricContextGetRepository.SetContext(app.MetricMemoryGetRepository)
		app.MetricContextListRepository.SetContext(app.MetricMemoryListRepository)
	}

	app.MetricUpdatesService = services.NewMetricUpdatesService(
		services.WithMetricUpdatesGetter(app.MetricContextGetRepository),
		services.WithMetricUpdatesSaver(app.MetricContextSaveRepository),
	)

	app.MetricGetService = services.NewMetricGetService(
		services.WithMetricGetGetter(app.MetricContextGetRepository),
	)

	app.MetricListService = services.NewMetricListService(
		services.WithMetricListLister(app.MetricContextListRepository),
	)

	app.Router = chi.NewRouter()

	app.MetricUpdatePathHandler = handlers.NewMetricUpdatePathHandler(handlers.WithMetricUpdaterPath(app.MetricUpdatesService))
	app.MetricUpdatePathHandler.RegisterRoute(app.Router)

	app.MetricUpdateBodyHandler = handlers.NewMetricUpdateBodyHandler(handlers.WithMetricUpdaterBody(app.MetricUpdatesService))
	app.MetricUpdateBodyHandler.RegisterRoute(app.Router)

	app.MetricUpdatesBodyHandler = handlers.NewMetricUpdatesBodyHandler(handlers.WithMetricUpdaterBatchBody(app.MetricUpdatesService))
	app.MetricUpdatesBodyHandler.RegisterRoute(app.Router)

	app.MetricGetPathHandler = handlers.NewMetricGetPathHandler(handlers.WithMetricGetterPath(app.MetricGetService))
	app.MetricGetPathHandler.RegisterRoute(app.Router)

	app.MetricGetBodyHandler = handlers.NewMetricGetBodyHandler(handlers.WithMetricGetterBody(app.MetricGetService))
	app.MetricGetBodyHandler.RegisterRoute(app.Router)

	app.MetricListHTMLHandler = handlers.NewMetricListHTMLHandler(handlers.WithMetricLister(app.MetricListService))
	app.MetricListHTMLHandler.RegisterRoute(app.Router)

	app.PingHandlerHandler = handlers.NewPingDBHandler(handlers.WithPingDB(app.DB))
	app.PingHandlerHandler.RegisterRoute(app.Router)

	app.Srv = &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: app.Router,
	}

	if app.Config.FileStoragePath != "" {
		app.Workers = append(
			app.Workers,
			workers.NewServerWorker(
				workers.WithRestore(cfg.Restore),
				workers.WithStoreInterval(cfg.StoreInterval),
				workers.WithLister(app.MetricContextListRepository),
				workers.WithSaver(app.MetricContextSaveRepository),
				workers.WithListerFile(app.MetricFileListRepository),
				workers.WithSaverFile(app.MetricFileSaveRepository),
			),
		)
	}

	return &app, nil
}

// Run starts the application server and its workers, handling graceful shutdown.
func (app *ServerApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	defer logger.Sync()

	errCh := make(chan error, 2)

	for _, worker := range app.Workers {
		go func(w func(context.Context) error) {
			logger.Log.Info("Worker goroutine started")
			errCh <- w(ctx)
		}(worker)
	}

	go func() {
		logger.Log.Infof("Starting HTTP server on %s", app.Config.ServerAddress)
		if err := app.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("HTTP server error: " + err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Error received: " + err.Error())
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Log.Info("Shutting down HTTP server gracefully")
	if err := app.Srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Error during server shutdown: " + err.Error())
		return err
	}

	logger.Log.Info("Server gracefully stopped")
	return nil
}
