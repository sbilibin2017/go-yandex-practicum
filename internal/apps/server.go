package apps

import (
	"context"
	"net"
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
	"google.golang.org/grpc"

	pb "github.com/sbilibin2017/go-yandex-practicum/protos"
)

// ServerAppConfig holds the configuration parameters for the server application.
type serverAppConfig struct {
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
type ServerAppOpt func(*serverAppConfig)

// NewServerAppConfig creates a new ServerAppConfig with the provided options.
func newServerAppConfig(opts ...ServerAppOpt) *serverAppConfig {
	cfg := &serverAppConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// WithServerAddress sets the server address (host:port) for the application.
func WithServerAddress(addr string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.ServerAddress = addr
	}
}

// WithServerDatabaseDSN sets the Data Source Name (DSN) for the database connection.
func WithServerDatabaseDSN(dsn string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.DatabaseDSN = dsn
	}
}

// WithServerStoreInterval sets the interval in seconds to store metrics data.
func WithServerStoreInterval(interval int) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.StoreInterval = interval
	}
}

// WithServerFileStoragePath sets the file path for storing metrics on disk.
func WithServerFileStoragePath(path string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.FileStoragePath = path
	}
}

// WithServerRestore enables or disables restoring metrics from file on startup.
func WithServerRestore(restore bool) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.Restore = restore
	}
}

// WithServerKey sets the application key used for hashing or encryption.
func WithServerKey(key string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.Key = key
	}
}

// WithServerCryptoKey sets the path to the private key used for encryption.
func WithServerCryptoKey(cryptoKey string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.CryptoKey = cryptoKey
	}
}

// WithServerConfigPath sets the path to the external configuration file.
func WithServerConfigPath(configPath string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.ConfigPath = configPath
	}
}

// WithServerTrustedSubnet sets the trusted subnet in CIDR notation.
func WithServerTrustedSubnet(subnet string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.TrustedSubnet = subnet
	}
}

// WithServerHashHeader sets the name of the HTTP header used to pass the SHA256 hash.
func WithServerHashHeader(hashHeader string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.HashHeader = hashHeader
	}
}

// WithServerLogLevel sets the logging level (e.g., debug, info, warn, error).
func WithServerLogLevel(logLevel string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.LogLevel = logLevel
	}
}

// WithServerMigrationsDir sets the directory path containing database migration files.
func WithServerMigrationsDir(migrationsDir string) ServerAppOpt {
	return func(c *serverAppConfig) {
		c.MigrationsDir = migrationsDir
	}
}

// [ServerAppOpt setters omitted for brevity: same as your code]

// ServerApp represents the main application server.
type ServerApp struct {
	Config    *serverAppConfig
	Container *container

	MetricUpdatePathHandler  *handlers.MetricUpdatePathHandler
	MetricUpdateBodyHandler  *handlers.MetricUpdateBodyHandler
	MetricUpdatesBodyHandler *handlers.MetricUpdatesBodyHandler
	MetricGetPathHandler     *handlers.MetricGetPathHandler
	MetricGetBodyHandler     *handlers.MetricGetBodyHandler
	MetricListHTMLHandler    *handlers.MetricListHTMLHandler

	PingHandlerHandler *handlers.PingDBHandler

	Router *chi.Mux
	Srv    *http.Server
}

func NewServerApp(opts ...ServerAppOpt) (*ServerApp, error) {
	cfg := newServerAppConfig(opts...)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Error("Failed to initialize logger: " + err.Error())
		return nil, err
	}

	app := &ServerApp{
		Config: cfg,
	}

	var err error
	app.Container, err = newContainer(cfg)
	if err != nil {
		return nil, err
	}

	app.Router = chi.NewRouter()

	// Initialize handlers with services from container
	app.MetricUpdatePathHandler = handlers.NewMetricUpdatePathHandler(
		handlers.WithMetricUpdaterPath(app.Container.MetricUpdatesService),
	)
	app.MetricUpdatePathHandler.RegisterRoute(app.Router)

	app.MetricUpdateBodyHandler = handlers.NewMetricUpdateBodyHandler(
		handlers.WithMetricUpdaterBody(app.Container.MetricUpdatesService),
	)
	app.MetricUpdateBodyHandler.RegisterRoute(app.Router)

	app.MetricUpdatesBodyHandler = handlers.NewMetricUpdatesBodyHandler(
		handlers.WithMetricUpdaterBatchBody(app.Container.MetricUpdatesService),
	)
	app.MetricUpdatesBodyHandler.RegisterRoute(app.Router)

	app.MetricGetPathHandler = handlers.NewMetricGetPathHandler(
		handlers.WithMetricGetterPath(app.Container.MetricGetService),
	)
	app.MetricGetPathHandler.RegisterRoute(app.Router)

	app.MetricGetBodyHandler = handlers.NewMetricGetBodyHandler(
		handlers.WithMetricGetterBody(app.Container.MetricGetService),
	)
	app.MetricGetBodyHandler.RegisterRoute(app.Router)

	app.MetricListHTMLHandler = handlers.NewMetricListHTMLHandler(
		handlers.WithMetricLister(app.Container.MetricListService),
	)
	app.MetricListHTMLHandler.RegisterRoute(app.Router)

	app.PingHandlerHandler = handlers.NewPingDBHandler(
		handlers.WithPingDB(app.Container.DB),
	)
	app.PingHandlerHandler.RegisterRoute(app.Router)

	app.Srv = &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: app.Router,
	}

	return app, nil
}

func (app *ServerApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	defer logger.Sync()

	errCh := make(chan error, 2)

	for _, worker := range app.Container.Workers {
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

type ServerGRPCApp struct {
	Config    *serverAppConfig
	Container *container

	MetricGRPCUpdaterHandler *handlers.MetricGRPCUpdaterHandler

	Server   *grpc.Server
	Listener net.Listener
}

// NewServerGRPCApp creates a new ServerGRPCApp.
func NewServerGRPCApp(opts ...ServerAppOpt) (*ServerGRPCApp, error) {
	cfg := newServerAppConfig(opts...)

	// Initialize logger
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return nil, err
	}

	// Create container with dependencies
	container, err := newContainer(cfg)
	if err != nil {
		return nil, err
	}

	app := &ServerGRPCApp{
		Config:    cfg,
		Container: container,
	}

	// Create handler with injected MetricUpdatesService
	app.MetricGRPCUpdaterHandler = handlers.NewMetricGRPCUpdaterHandler(container.MetricUpdatesService)

	app.Listener, err = net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		return nil, err
	}

	app.Server = grpc.NewServer()
	pb.RegisterMetricUpdaterServer(app.Server, app.MetricGRPCUpdaterHandler)

	return app, nil
}

// Run starts the gRPC server and workers, handling graceful shutdown.
func (app *ServerGRPCApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	defer logger.Sync()

	errCh := make(chan error, 2)

	// Start workers
	for _, worker := range app.Container.Workers {
		go func(w func(context.Context) error) {
			logger.Log.Info("Worker goroutine started")
			errCh <- w(ctx)
		}(worker)
	}

	// Start gRPC server
	go func() {
		logger.Log.Infof("Starting gRPC server on %s", app.Config.ServerAddress)
		if err := app.Server.Serve(app.Listener); err != nil {
			logger.Log.Error("gRPC server error: " + err.Error())
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

	// Graceful shutdown
	done := make(chan struct{})
	go func() {
		app.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("Timeout on gRPC server graceful stop, forcing stop")
		app.Server.Stop()
	}

	logger.Log.Info("Server stopped")
	return nil
}

// Container holds dependencies.
type container struct {
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

	Workers []func(ctx context.Context) error
}

func newContainer(cfg *serverAppConfig) (*container, error) {
	c := &container{}

	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return nil, err
		}

		if cfg.MigrationsDir != "" {
			if err := goose.SetDialect("postgres"); err != nil {
				return nil, err
			}
			if err := goose.Up(db.DB, cfg.MigrationsDir); err != nil {
				return nil, err
			}
		}

		logger.Log.Info("DB initialized and migrations applied")

		c.DB = db
		c.MetricDBSaveRepository = repositories.NewMetricDBSaveRepository(
			repositories.WithMetricDBSaveRepositoryDB(db),
			repositories.WithMetricDBSaveRepositoryTxGetter(contexts.GetTxFromContext),
		)
		c.MetricDBGetRepository = repositories.NewMetricDBGetRepository(
			repositories.WithMetricDBGetRepositoryDB(db),
			repositories.WithMetricDBGetRepositoryTxGetter(contexts.GetTxFromContext),
		)
		c.MetricDBListRepository = repositories.NewMetricDBListRepository(
			repositories.WithMetricDBListRepositoryDB(db),
			repositories.WithMetricDBListRepositoryTxGetter(contexts.GetTxFromContext),
		)
	}

	if cfg.DatabaseDSN == "" && cfg.FileStoragePath != "" {
		logger.Log.Infof("Using file storage at: %s", cfg.FileStoragePath)
		if err := os.MkdirAll(filepath.Dir(cfg.FileStoragePath), 0755); err != nil {
			return nil, err
		}

		c.MetricFileSaveRepository = repositories.NewMetricFileSaveRepository(
			repositories.WithMetricFileSaveRepositoryPath(cfg.FileStoragePath),
		)
		c.MetricFileGetRepository = repositories.NewMetricFileGetRepository(
			repositories.WithMetricFileGetRepositoryPath(cfg.FileStoragePath),
		)
		c.MetricFileListRepository = repositories.NewMetricFileListRepository(
			repositories.WithMetricFileListRepositoryPath(cfg.FileStoragePath),
		)
	}

	if cfg.DatabaseDSN == "" && cfg.FileStoragePath == "" {
		logger.Log.Info("Using in-memory metric storage")
		c.MetricMemorySaveRepository = repositories.NewMetricMemorySaveRepository()
		c.MetricMemoryGetRepository = repositories.NewMetricMemoryGetRepository()
		c.MetricMemoryListRepository = repositories.NewMetricMemoryListRepository()
	}

	// Context repositories always initialized
	c.MetricContextSaveRepository = repositories.NewMetricContextSaveRepository()
	c.MetricContextGetRepository = repositories.NewMetricContextGetRepository()
	c.MetricContextListRepository = repositories.NewMetricContextListRepository()

	switch {
	case c.MetricDBSaveRepository != nil:
		c.MetricContextSaveRepository.SetContext(c.MetricDBSaveRepository)
		c.MetricContextGetRepository.SetContext(c.MetricDBGetRepository)
		c.MetricContextListRepository.SetContext(c.MetricDBListRepository)

	case c.MetricFileSaveRepository != nil:
		c.MetricContextSaveRepository.SetContext(c.MetricFileSaveRepository)
		c.MetricContextGetRepository.SetContext(c.MetricFileGetRepository)
		c.MetricContextListRepository.SetContext(c.MetricFileListRepository)

	default:
		c.MetricContextSaveRepository.SetContext(c.MetricMemorySaveRepository)
		c.MetricContextGetRepository.SetContext(c.MetricMemoryGetRepository)
		c.MetricContextListRepository.SetContext(c.MetricMemoryListRepository)
	}

	c.MetricUpdatesService = services.NewMetricUpdatesService(
		services.WithMetricUpdatesGetter(c.MetricContextGetRepository),
		services.WithMetricUpdatesSaver(c.MetricContextSaveRepository),
	)
	c.MetricGetService = services.NewMetricGetService(
		services.WithMetricGetGetter(c.MetricContextGetRepository),
	)
	c.MetricListService = services.NewMetricListService(
		services.WithMetricListLister(c.MetricContextListRepository),
	)

	if cfg.FileStoragePath != "" {
		c.Workers = append(
			c.Workers,
			workers.NewServerWorker(
				workers.WithRestore(cfg.Restore),
				workers.WithStoreInterval(cfg.StoreInterval),
				workers.WithLister(c.MetricContextListRepository),
				workers.WithSaver(c.MetricContextSaveRepository),
				workers.WithListerFile(c.MetricFileListRepository),
				workers.WithSaverFile(c.MetricFileSaveRepository),
			),
		)
	}

	return c, nil
}
