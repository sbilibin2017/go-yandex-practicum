package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	flagAddress        = "a"
	flagUpdateEndpoint = "u"
	flagPollInterval   = "p"
	flagReportInterval = "r"
	flagLogLevel       = "l"

	envAddress        = "ADDRESS"
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envUpdateEndpoint = "UPDATE_ENDPOINT"
	envLogLevel       = "LOG_LEVEL"

	defaultServerAddress  = "http://localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultUpdateEndpoint = "/update/"
	defaultLogLevel       = "info"

	flagAddressUsage        = "Metrics server address"
	flagPollIntervalUsage   = "Poll interval in seconds"
	flagReportIntervalUsage = "Report interval in seconds"
	flagLogLevelUsage       = "Log level (e.g., debug, info, warn, error)"
	flagUpdateEndpointUsage = "Metrics server update endpoint"
)

type options struct {
	ServerAddress        string
	ServerUpdateEndpoint string
	PollInterval         int
	ReportInterval       int
	LogLevel             string
}

var opts options

func parseFlags() *options {
	flag.StringVar(&opts.ServerAddress, flagAddress, defaultServerAddress, flagAddressUsage)
	flag.StringVar(&opts.ServerUpdateEndpoint, flagUpdateEndpoint, defaultUpdateEndpoint, flagUpdateEndpointUsage)
	flag.IntVar(&opts.PollInterval, flagPollInterval, defaultPollInterval, flagPollIntervalUsage)
	flag.IntVar(&opts.ReportInterval, flagReportInterval, defaultReportInterval, flagReportIntervalUsage)
	flag.StringVar(&opts.LogLevel, flagLogLevel, defaultLogLevel, flagLogLevelUsage)

	flag.Parse()

	if envServerAddress := os.Getenv(envAddress); envServerAddress != "" {
		opts.ServerAddress = envServerAddress
	}
	if envServerUpdateEndpoint := os.Getenv(envUpdateEndpoint); envServerUpdateEndpoint != "" {
		opts.ServerUpdateEndpoint = envServerUpdateEndpoint
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

	return &opts
}
