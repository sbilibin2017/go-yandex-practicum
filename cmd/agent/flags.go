package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	flagAddress        = "a"
	flagPollInterval   = "p"
	flagReportInterval = "r"
	flagLogLevel       = "ll"
	flagKey            = "k"
	flagRateLimit      = "l"
	flagNumWorkers     = "w"

	envAddress        = "ADDRESS"
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envLogLevel       = "LOG_LEVEL"
	envKey            = "KEY"
	envRateLimit      = "RATE_LIMIT"
	envNumWorkers     = "NUM_WORKERS"

	defaultServerAddress  = "http://localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultLogLevel       = "info"
	defaultNumWorkers     = 5

	flagAddressUsage        = "Metrics server address"
	flagPollIntervalUsage   = "Poll interval in seconds"
	flagReportIntervalUsage = "Report interval in seconds"
	flagLogLevelUsage       = "Log level (e.g., debug, info, warn, error)"
	flagKeyUsage            = "Key for HMAC SHA256 hash"
	flagRateLimitUsage      = "Max number of concurrent outgoing requests"
	flagNumWorkersUsage     = "Number of workers"
)

type options struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
	LogLevel       string
	Key            string
	RateLimit      int
	NumWorkers     int
}

var opts options

func parseFlags() *options {
	flag.StringVar(&opts.ServerAddress, flagAddress, defaultServerAddress, flagAddressUsage)
	flag.IntVar(&opts.PollInterval, flagPollInterval, defaultPollInterval, flagPollIntervalUsage)
	flag.IntVar(&opts.ReportInterval, flagReportInterval, defaultReportInterval, flagReportIntervalUsage)
	flag.StringVar(&opts.LogLevel, flagLogLevel, defaultLogLevel, flagLogLevelUsage)
	flag.StringVar(&opts.Key, flagKey, "", flagKeyUsage)
	flag.IntVar(&opts.RateLimit, flagRateLimit, 0, flagRateLimitUsage)
	flag.IntVar(&opts.NumWorkers, flagNumWorkers, defaultNumWorkers, flagNumWorkersUsage)

	flag.Parse()

	if envServerAddress := os.Getenv(envAddress); envServerAddress != "" {
		opts.ServerAddress = envServerAddress
	}
	if envPollInterval := os.Getenv(envPollInterval); envPollInterval != "" {
		if val, err := strconv.Atoi(envPollInterval); err == nil {
			opts.PollInterval = val
		}
	}
	if envReportInterval := os.Getenv(envReportInterval); envReportInterval != "" {
		if val, err := strconv.Atoi(envReportInterval); err == nil {
			opts.ReportInterval = val
		}
	}
	if envLogLevel := os.Getenv(envLogLevel); envLogLevel != "" {
		opts.LogLevel = envLogLevel
	}
	if envKey := os.Getenv(envKey); envKey != "" {
		opts.Key = envKey
	}
	if envRateLimit := os.Getenv(envRateLimit); envRateLimit != "" {
		if val, err := strconv.Atoi(envRateLimit); err == nil {
			opts.RateLimit = val
		}
	}
	if envNumWorkers := os.Getenv(envNumWorkers); envNumWorkers != "" {
		if val, err := strconv.Atoi(envNumWorkers); err == nil {
			opts.NumWorkers = val
		}
	}

	return &opts
}
