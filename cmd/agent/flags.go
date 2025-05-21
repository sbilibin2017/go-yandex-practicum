package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	flagAddressName        = "a"
	flagPollIntervalName   = "p"
	flagReportIntervalName = "r"
	flagKeyName            = "k"
	flagRateLimitName      = "l"

	envAddress        = "ADDRESS"
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envKey            = "KEY"
	envRateLimit      = "RATE_LIMIT"

	defaultAddress        = "http://localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultKey            = ""
	defaultRateLimit      = 0

	descAddress        = "Metrics server address"
	descPollInterval   = "Poll interval in seconds"
	descReportInterval = "Report interval in seconds"
	descKey            = "Key for HMAC SHA256 hash"
	descRateLimit      = "Max number of concurrent outgoing requests"
)

const (
	batchSize     = 100
	logLevel      = "info"
	hashKeyHeader = "HashSHA256"
	emptyString   = ""
)

var (
	flagServerAddress  string
	flagPollInterval   int
	flagReportInterval int
	flagKey            string
	flagRateLimit      int
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, flagAddressName, defaultAddress, descAddress)
	flag.IntVar(&flagPollInterval, flagPollIntervalName, defaultPollInterval, descPollInterval)
	flag.IntVar(&flagReportInterval, flagReportIntervalName, defaultReportInterval, descReportInterval)
	flag.StringVar(&flagKey, flagKeyName, defaultKey, descKey)
	flag.IntVar(&flagRateLimit, flagRateLimitName, defaultRateLimit, descRateLimit)

	flag.Parse()

	if env := os.Getenv(envAddress); env != emptyString {
		flagServerAddress = env
	}
	if env := os.Getenv(envPollInterval); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagPollInterval = v
		}
	}
	if env := os.Getenv(envReportInterval); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagReportInterval = v
		}
	}
	if env := os.Getenv(envKey); env != emptyString {
		flagKey = env
	}
	if env := os.Getenv(envRateLimit); env != emptyString {
		if v, err := strconv.Atoi(env); err == nil {
			flagRateLimit = v
		}
	}
}
