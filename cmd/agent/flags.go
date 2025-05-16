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
	flagKey            string
	flagRateLimit      int
	flagNumWorkers     int
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, "a", "http://localhost:8080", "Metrics server address")
	flag.IntVar(&flagPollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&flagReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&flagLogLevel, "ll", "info", "Log level (e.g., debug, info, warn, error)")
	flag.StringVar(&flagKey, "k", "", "Key for HMAC SHA256 hash")
	flag.IntVar(&flagRateLimit, "l", 0, "Max number of concurrent outgoing requests")
	flag.IntVar(&flagNumWorkers, "w", 5, "Number of workers")

	flag.Parse()

	if val := os.Getenv("ADDRESS"); val != "" {
		flagServerAddress = val
	}
	if val := os.Getenv("POLL_INTERVAL"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			flagPollInterval = v
		}
	}
	if val := os.Getenv("REPORT_INTERVAL"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			flagReportInterval = v
		}
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		flagLogLevel = val
	}
	if val := os.Getenv("KEY"); val != "" {
		flagKey = val
	}
	if val := os.Getenv("RATE_LIMIT"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			flagRateLimit = v
		}
	}
	if val := os.Getenv("NUM_WORKERS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			flagNumWorkers = v
		}
	}
}
