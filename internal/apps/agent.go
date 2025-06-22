package apps

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

// AgentAppConfig holds the configuration parameters for the Agent application.
type agentAppConfig struct {
	ServerAddress  string // address and port where the agent server runs
	Header         string // HTTP header used for requests
	PollInterval   int    // how often (seconds) the agent polls for updates
	ReportInterval int    // how often (seconds) the agent reports data
	Key            string // secret key used for signing or encryption
	RateLimit      int    // maximum rate of requests allowed
	CryptoKey      string // path to private key file used for encryption
	ConfigPath     string // path to the config file
	Restore        bool   // whether to restore data from backup on startup
	HashHeader     string // HTTP header key for the SHA256 hash
	LogLevel       string // logging verbosity level (e.g., debug, info)
	BatchSize      int    // number of items processed in a batch
	IsGRPC         bool
}

// AgentAppOpt represents a functional option for configuring the AgentAppConfig.
type AgentAppOpt func(*agentAppConfig)

// NewAgentAppConfig creates a new AgentAppConfig using the provided functional options.
func newAgentAppConfig(opts ...AgentAppOpt) *agentAppConfig {
	cfg := &agentAppConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// WithAgentServerAddress sets the server address in the AgentAppConfig.
func WithAgentServerAddress(addr string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.ServerAddress = addr
	}
}

// WithAgentHeader sets the HTTP header in the AgentAppConfig.
func WithAgentHeader(header string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.Header = header
	}
}

// WithAgentPollInterval sets the polling interval in seconds.
func WithAgentPollInterval(interval int) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.PollInterval = interval
	}
}

// WithAgentReportInterval sets the report interval in seconds.
func WithAgentReportInterval(interval int) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.ReportInterval = interval
	}
}

// WithAgentKey sets the secret key used for signing or encryption.
func WithAgentKey(key string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.Key = key
	}
}

// WithAgentRateLimit sets the request rate limit.
func WithAgentRateLimit(rateLimit int) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.RateLimit = rateLimit
	}
}

// WithAgentCryptoKey sets the path to the public key file used for encryption.
func WithAgentCryptoKey(cryptoKey string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.CryptoKey = cryptoKey
	}
}

// WithAgentConfigPath sets the path to the configuration file.
func WithAgentConfigPath(path string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.ConfigPath = path
	}
}

// WithAgentRestore enables or disables restoration from backup.
func WithAgentRestore(restore bool) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.Restore = restore
	}
}

// WithAgentHashHeader sets the SHA256 hash HTTP header.
func WithAgentHashHeader(hashHeader string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.HashHeader = hashHeader
	}
}

// WithAgentLogLevel sets the log level for the application.
func WithAgentLogLevel(logLevel string) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.LogLevel = logLevel
	}
}

// WithAgentBatchSize sets the batch size for reporting metrics.
func WithAgentBatchSize(batchSize int) AgentAppOpt {
	return func(c *agentAppConfig) {
		c.BatchSize = batchSize
	}
}

// WithAgentBatchSize sets the batch size for reporting metrics.
func WithGRPC() AgentAppOpt {
	return func(c *agentAppConfig) {
		c.IsGRPC = true
	}
}

// AgentApp is the main struct representing the agent application.
type AgentApp struct {
	Config              *agentAppConfig
	MetricContextFacade *facades.MetricFacadeContext
	Workers             []func(ctx context.Context) error
}

// NewAgentApp creates and initializes a new AgentApp with the provided options.
func NewAgentApp(opts ...AgentAppOpt) (*AgentApp, error) {
	config := newAgentAppConfig(opts...)

	var app AgentApp
	app.Config = config

	err := logger.Initialize(config.LogLevel)
	if err != nil {
		return nil, err
	}

	app.MetricContextFacade = facades.NewMetricFacadeContext()

	if !config.IsGRPC {
		metricFacade, err := facades.NewMetricHTTPFacade(
			facades.WithMetricFacadeServerAddress(config.ServerAddress),
			facades.WithMetricFacadeHeader(config.Header),
			facades.WithMetricFacadeKey(config.Key),
			facades.WithMetricFacadeCryptoKeyPath(config.CryptoKey),
		)
		if err != nil {
			logger.Log.Error("Failed to create MetricFacade:", err)
			return nil, err
		}
		app.MetricContextFacade.SetContext(metricFacade)
	} else {
		metricFacade, err := facades.NewMetricGRPCFacade(
			facades.WithMetricGRPCServerAddress(config.ServerAddress),
		)
		if err != nil {
			logger.Log.Error("Failed to create MetricFacade:", err)
			return nil, err
		}
		app.MetricContextFacade.SetContext(metricFacade)
	}

	app.Workers = append(
		app.Workers,
		workers.NewAgentWorker(
			workers.WithPollInterval(config.PollInterval),
			workers.WithReportInterval(config.ReportInterval),
			workers.WithBatchSize(config.BatchSize),
			workers.WithRateLimit(config.RateLimit),
			workers.WithUpdater(app.MetricContextFacade),
		),
	)

	return &app, nil
}

// Run starts the AgentApp and waits for shutdown signals.
// It listens for SIGINT, SIGTERM, or SIGQUIT and gracefully shuts down all workers.
func (app *AgentApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	defer logger.Sync()

	errCh := make(chan error, len(app.Workers))

	for _, worker := range app.Workers {
		go func(w func(context.Context) error) {
			logger.Log.Info("Worker goroutine started")
			errCh <- w(ctx)
		}(worker)
	}

	select {
	case err := <-errCh:
		if err != nil {
			logger.Log.Error("Worker returned error:", err)
		}
		return err
	case <-ctx.Done():
		logger.Log.Info("Received termination signal, shutting down")
		return ctx.Err()
	}
}
