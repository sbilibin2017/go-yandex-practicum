package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/configs"
	"github.com/spf13/pflag"
)

const (
	hashKeyHeader = "HashSHA256"
	logLevel      = "info"
	batchSize     = 100
)

var (
	flagServerAddress  string
	flagHeader         string
	flagPollInterval   int
	flagReportInterval int
	flagKey            string
	flagRateLimit      int
	flagCryptoKey      string
	flagBatchSize      int
	flagConfigPath     string
)

func flags() error {
	if err := parseFlags(); err != nil {
		return err
	}
	if err := parseConfigFile(); err != nil {
		return err
	}
	if err := parseEnvs(); err != nil {
		return err
	}
	return nil
}

func parseFlags() error {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	pflag.StringVarP(&flagServerAddress, "address", "a", "http://localhost:8080", "Metrics server address")
	pflag.StringVar(&flagHeader, "header", "", "Header name for metrics hash")
	pflag.IntVarP(&flagPollInterval, "poll-interval", "p", 2, "Poll interval in seconds")
	pflag.IntVarP(&flagReportInterval, "report-interval", "r", 10, "Report interval in seconds")
	pflag.StringVarP(&flagKey, "key", "k", "", "Key for HMAC SHA256 hash")
	pflag.IntVarP(&flagRateLimit, "rate-limit", "l", 0, "Max number of concurrent outgoing requests")
	pflag.IntVar(&flagBatchSize, "batch-size", batchSize, "Batch size for metrics reporting")
	pflag.StringVar(&flagCryptoKey, "crypto-key", "", "Path to public key file for encryption")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "Path to config file")

	return pflag.CommandLine.Parse(os.Args[1:])
}

func parseEnvs() error {
	if env := os.Getenv("ADDRESS"); env != "" {
		flagServerAddress = env
	}
	if env := os.Getenv("HEADER"); env != "" {
		flagHeader = env
	}
	if env := os.Getenv("POLL_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			flagPollInterval = v
		}
	}
	if env := os.Getenv("REPORT_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			flagReportInterval = v
		}
	}
	if env := os.Getenv("KEY"); env != "" {
		flagKey = env
	}
	if env := os.Getenv("RATE_LIMIT"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			flagRateLimit = v
		}
	}
	if env := os.Getenv("BATCH_SIZE"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			flagBatchSize = v
		}
	}
	if env := os.Getenv("CRYPTO_KEY"); env != "" {
		flagCryptoKey = env
	}
	if env := os.Getenv("CONFIG"); env != "" {
		flagConfigPath = env
	}
	return nil
}

func parseConfigFile() error {
	if flagConfigPath == "" {
		return nil
	}

	file, err := os.Open(flagConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var cfg configs.AgentConfig

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}

	if cfg.ServerAddress != nil {
		flagServerAddress = *cfg.ServerAddress
	}
	if cfg.Header != nil {
		flagHeader = *cfg.Header
	}
	if cfg.ReportInterval != nil {
		flagReportInterval = *cfg.ReportInterval
	}
	if cfg.PollInterval != nil {
		flagPollInterval = *cfg.PollInterval
	}
	if cfg.Key != nil {
		flagKey = *cfg.Key
	}
	if cfg.RateLimit != nil {
		flagRateLimit = *cfg.RateLimit
	}
	if cfg.BatchSize != nil {
		flagBatchSize = *cfg.BatchSize
	}
	if cfg.CryptoKeyPath != nil {
		flagCryptoKey = *cfg.CryptoKeyPath
	}

	return nil
}

func parseDurationToSeconds(dur string) (int, error) {
	if seconds, err := strconv.Atoi(dur); err == nil {
		return seconds, nil
	}
	parsed, err := time.ParseDuration(dur)
	if err != nil {
		return 0, err
	}
	return int(parsed.Seconds()), nil
}
