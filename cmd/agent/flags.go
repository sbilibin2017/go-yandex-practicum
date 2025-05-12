package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagServerAddress  string
	flagPollInterval   int
	flagReportInterval int
	flagLogLevel       string
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, "a", "http://localhost:8080", "Metrics server address")
	flag.IntVar(&flagPollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&flagReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		flagServerAddress = envServerAddress
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := strconv.Atoi(envPoll); err == nil {
			flagPollInterval = val
		}
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := strconv.Atoi(envReport); err == nil {
			flagReportInterval = val
		}
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
}
