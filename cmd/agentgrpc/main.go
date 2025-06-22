package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/apps"
	"github.com/spf13/pflag"
)

// main is the program entry point.
//
// It prints build information, parses command-line flags, config file, and environment variables,
// and then starts the agent application with the parsed configuration.
// The program panics if any critical error occurs during setup or runtime.
func main() {
	printBuildInfo()

	if err := parseFlags(); err != nil {
		panic(err)
	}

	if err := parseConfigFile(); err != nil {
		panic(err)
	}

	if err := parseEnvs(); err != nil {
		panic(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

// Build metadata variables set via build flags (ldflags).
var (
	buildVersion string // application version
	buildDate    string // build timestamp
	buildCommit  string // Git commit hash
)

// Configuration flags populated by flags, env vars, and config file.
var (
	flagServerAddress  string // metrics server address
	flagPollInterval   int    // poll interval in seconds
	flagReportInterval int    // report interval in seconds
	flagRateLimit      int    // max number of concurrent outgoing requests
	flagConfigPath     string // path to JSON config file
	flagRestore        bool   // whether to restore data from backup
	flagLogLevel       string // application log level
	flagBatchSize      int    // batch size for metrics reporting
)

// parseFlags parses command-line flags and stores their values in the global config variables.
func parseFlags() error {
	pflag.StringVarP(&flagServerAddress, "address", "a", "http://localhost:8080", "Metrics server address")
	pflag.IntVarP(&flagPollInterval, "poll-interval", "p", 2, "Poll interval in seconds")
	pflag.IntVarP(&flagReportInterval, "report-interval", "r", 10, "Report interval in seconds")
	pflag.IntVarP(&flagRateLimit, "rate-limit", "l", 0, "Max number of concurrent outgoing requests")
	pflag.BoolVar(&flagRestore, "restore", false, "Whether to restore data from backup")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "Path to config file")
	pflag.StringVarP(&flagLogLevel, "log-level", "L", "info", "Log level for the application")
	pflag.IntVarP(&flagBatchSize, "batch-size", "b", 100, "Batch size for metrics reporting")

	pflag.Parse()
	return nil
}

// parseConfigFile reads configuration from a JSON file (if provided).
// It overrides previously parsed flags with values found in the config file.
func parseConfigFile() error {
	if flagConfigPath == "" {
		return nil
	}

	if _, err := os.Stat(flagConfigPath); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(flagConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	cfg := &struct {
		ServerAddress  *string `json:"server_address,omitempty"`
		PollInterval   *int    `json:"poll_interval,omitempty"`
		ReportInterval *int    `json:"report_interval,omitempty"`
		RateLimit      *int    `json:"rate_limit,omitempty"`
		Restore        *bool   `json:"restore,omitempty"`
		LogLevel       *string `json:"log_level,omitempty"`
		BatchSize      *int    `json:"batch_size,omitempty"`
	}{}

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return err
	}

	if cfg.ServerAddress != nil {
		flagServerAddress = *cfg.ServerAddress
	}
	if cfg.PollInterval != nil {
		flagPollInterval = *cfg.PollInterval
	}
	if cfg.ReportInterval != nil {
		flagReportInterval = *cfg.ReportInterval
	}
	if cfg.RateLimit != nil {
		flagRateLimit = *cfg.RateLimit
	}
	if cfg.Restore != nil {
		flagRestore = *cfg.Restore
	}
	if cfg.LogLevel != nil {
		flagLogLevel = *cfg.LogLevel
	}
	if cfg.BatchSize != nil {
		flagBatchSize = *cfg.BatchSize
	}

	return nil
}

// parseEnvs loads configuration values from environment variables.
// If a variable is set, it overrides the previously set configuration.
func parseEnvs() error {
	if v := os.Getenv("ADDRESS"); v != "" {
		flagServerAddress = v
	}
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			flagPollInterval = val
		}
	}
	if v := os.Getenv("REPORT_INTERVAL"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			flagReportInterval = val
		}
	}
	if v := os.Getenv("RATE_LIMIT"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			flagRateLimit = val
		}
	}
	if v := os.Getenv("RESTORE"); v != "" {
		if val, err := strconv.ParseBool(v); err == nil {
			flagRestore = val
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		flagLogLevel = v
	}
	if v := os.Getenv("BATCH_SIZE"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			flagBatchSize = val
		}
	}

	return nil
}

// run initializes and runs the agent application using the parsed configuration.
func run() error {
	app, err := apps.NewAgentApp(
		apps.WithAgentServerAddress(flagServerAddress),
		apps.WithAgentPollInterval(flagPollInterval),
		apps.WithAgentReportInterval(flagReportInterval),
		apps.WithAgentBatchSize(flagBatchSize),
		apps.WithAgentRateLimit(flagRateLimit),
		apps.WithAgentLogLevel(flagLogLevel),
		apps.WithGRPC(),
	)

	if err != nil {
		return err
	}

	return app.Run(context.Background())
}

// printBuildInfo prints versioning and build metadata to stdout.
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
