package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

const (
	hashKeyHeader = "HashSHA256"
	logLevel      = "info"
)

var (
	flagServerAddress   string
	flagDatabaseDSN     string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagKey             string
	flagCryptoKey       string
	flagConfigPath      string
	flagTrustedSubnet   string
)

func init() {
	parseFlags()
	err := parseConfigFile()
	if err != nil {
		panic(err)
	}
	parseEnv()
}

func parseFlags() {
	pflag.StringVarP(&flagServerAddress, "address", "a", ":8080", "address and port to run server")
	pflag.StringVarP(&flagDatabaseDSN, "dsn", "d", "", "dsn for database connection")
	pflag.IntVarP(&flagStoreInterval, "interval", "i", 300, "interval (in seconds) to store data")
	pflag.StringVarP(&flagFileStoragePath, "file", "f", "", "path to store files")
	pflag.BoolVarP(&flagRestore, "restore", "r", false, "whether to restore data from backup")
	pflag.StringVarP(&flagKey, "key", "k", "", "key used for SHA256 hashing")
	pflag.StringVar(&flagCryptoKey, "crypto-key", "", "path to private key file for encryption")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "path to config file")
	pflag.StringVarP(&flagTrustedSubnet, "trusted-subnet", "t", "", "trusted subnet in CIDR notation")

	pflag.Parse()
}

func parseEnv() {
	if env := os.Getenv("ADDRESS"); env != "" {
		flagServerAddress = env
	}
	if env := os.Getenv("DATABASE_DSN"); env != "" {
		flagDatabaseDSN = env
	}
	if env := os.Getenv("STORE_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			flagStoreInterval = v
		}
	}
	if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
		flagFileStoragePath = env
	}
	if env := os.Getenv("RESTORE"); env != "" {
		if v, err := strconv.ParseBool(env); err == nil {
			flagRestore = v
		}
	}
	if env := os.Getenv("KEY"); env != "" {
		flagKey = env
	}
	if env := os.Getenv("CRYPTO_KEY"); env != "" {
		flagCryptoKey = env
	}
	if env := os.Getenv("CONFIG"); env != "" {
		flagConfigPath = env
	}
	if env := os.Getenv("TRUSTED_SUBNET"); env != "" {
		flagTrustedSubnet = env
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
		Address       string `json:"address"`
		DatabaseDSN   string `json:"database_dsn"`
		StoreInterval int    `json:"store_interval"`
		FilePath      string `json:"store_file"`
		Restore       bool   `json:"restore"`
		Key           string `json:"key"`
		CryptoKey     string `json:"crypto_key"`
		TrustedSubnet string `json:"trusted_subnet"`
	}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}

	if cfg.Address != "" && flagServerAddress == ":8080" {
		flagServerAddress = cfg.Address
	}
	if cfg.DatabaseDSN != "" && flagDatabaseDSN == "" {
		flagDatabaseDSN = cfg.DatabaseDSN
	}
	if cfg.StoreInterval != 0 && flagStoreInterval == 300 {
		flagStoreInterval = cfg.StoreInterval
	}
	if cfg.FilePath != "" && flagFileStoragePath == "" {
		flagFileStoragePath = cfg.FilePath
	}
	if !flagRestore {
		flagRestore = cfg.Restore
	}
	if cfg.Key != "" && flagKey == "" {
		flagKey = cfg.Key
	}
	if cfg.CryptoKey != "" && flagCryptoKey == "" {
		flagCryptoKey = cfg.CryptoKey
	}

	if cfg.TrustedSubnet != "" && flagTrustedSubnet == "" {
		flagTrustedSubnet = cfg.TrustedSubnet
	}

	return nil
}
