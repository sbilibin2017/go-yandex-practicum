package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

const (
	hashKeyHeader = "HashSHA256"
	logLevel      = "info"
	batchSize     = 100
)

var (
	flagServerAddress  string
	flagPollInterval   int
	flagReportInterval int
	flagKey            string
	flagRateLimit      int
	flagCryptoKey      string
	flagConfigPath     string
)

func init() {
	parseFlags()
	if err := parseConfigFile(); err != nil {
		panic(err)
	}
	parseEnv()
}

func parseFlags() {
	pflag.StringVarP(&flagServerAddress, "address", "a", "http://localhost:8080", "Metrics server address")
	pflag.IntVarP(&flagPollInterval, "poll-interval", "p", 2, "Poll interval in seconds")
	pflag.IntVarP(&flagReportInterval, "report-interval", "r", 10, "Report interval in seconds")
	pflag.StringVarP(&flagKey, "key", "k", "", "Key for HMAC SHA256 hash")
	pflag.IntVarP(&flagRateLimit, "rate-limit", "l", 0, "Max number of concurrent outgoing requests")
	pflag.StringVar(&flagCryptoKey, "crypto-key", "", "Path to public key file for encryption")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "Path to config file")

	pflag.Parse()
}

func parseEnv() {
	if env := os.Getenv("ADDRESS"); env != "" {
		flagServerAddress = env
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
	if env := os.Getenv("CRYPTO_KEY"); env != "" {
		flagCryptoKey = env
	}
	if env := os.Getenv("CONFIG"); env != "" {
		flagConfigPath = env
	}
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

	var cfg struct {
		Address        string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		Key            string `json:"key"`
		RateLimit      int    `json:"rate_limit"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}

	if cfg.Address != "" && flagServerAddress == "http://localhost:8080" {
		flagServerAddress = cfg.Address
	}
	if cfg.ReportInterval != "" && flagReportInterval == 10 {
		if d, err := parseDurationToSeconds(cfg.ReportInterval); err == nil {
			flagReportInterval = d
		}
	}
	if cfg.PollInterval != "" && flagPollInterval == 2 {
		if d, err := parseDurationToSeconds(cfg.PollInterval); err == nil {
			flagPollInterval = d
		}
	}
	if cfg.Key != "" && flagKey == "" {
		flagKey = cfg.Key
	}
	if cfg.RateLimit != 0 && flagRateLimit == 0 {
		flagRateLimit = cfg.RateLimit
	}
	if cfg.CryptoKey != "" && flagCryptoKey == "" {
		flagCryptoKey = cfg.CryptoKey
	}

	return nil
}

func parseDurationToSeconds(dur string) (int, error) {
	// Support both simple integer seconds or duration strings like "10s"
	if seconds, err := strconv.Atoi(dur); err == nil {
		return seconds, nil
	}
	parsed, err := time.ParseDuration(dur)
	if err != nil {
		return 0, err
	}
	return int(parsed.Seconds()), nil
}
