package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	serverAddress  string
	pollInterval   int
	reportInterval int
	key            string
	rateLimit      int
	batchSize      int
	logLevel       string
	header         string
)

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "http://localhost:8080", "Metrics server address")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&key, "k", "", "Key for HMAC SHA256 hash")
	flag.IntVar(&rateLimit, "l", 0, "Max number of concurrent outgoing requests")

	flag.Parse()

	if env := os.Getenv("ADDRESS"); env != "" {
		serverAddress = env
	}
	if env := os.Getenv("POLL_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			pollInterval = v
		}
	}
	if env := os.Getenv("REPORT_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			reportInterval = v
		}
	}
	if env := os.Getenv("KEY"); env != "" {
		key = env
	}
	if env := os.Getenv("RATE_LIMIT"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			rateLimit = v
		}
	}

	logLevel = "info"
	header = "HashSHA256"
	batchSize = 100
}
