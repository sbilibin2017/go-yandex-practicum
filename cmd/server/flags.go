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

func flags() error {
	err := parseFlags()
	if err != nil {
		return err
	}

	err = parseConfigFile()
	if err != nil {
		return err
	}

	err = parseEnvs()
	if err != nil {
		return err
	}

	return nil
}

func parseFlags() error {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	pflag.StringVarP(&flagServerAddress, "address", "a", ":8080", "address and port to run server")
	pflag.StringVarP(&flagDatabaseDSN, "dsn", "d", "", "dsn for database connection")
	pflag.IntVarP(&flagStoreInterval, "interval", "i", 300, "interval (in seconds) to store data")
	pflag.StringVarP(&flagFileStoragePath, "file", "f", "", "path to store files")
	pflag.BoolVarP(&flagRestore, "restore", "r", false, "whether to restore data from backup")
	pflag.StringVarP(&flagKey, "key", "k", "", "key used for SHA256 hashing")
	pflag.StringVar(&flagCryptoKey, "crypto-key", "", "path to private key file for encryption")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "path to config file")
	pflag.StringVarP(&flagTrustedSubnet, "trusted-subnet", "t", "", "trusted subnet in CIDR notation")

	return pflag.CommandLine.Parse(os.Args[1:])
}

func parseEnvs() error {
	if v := os.Getenv("ADDRESS"); v != "" {
		flagServerAddress = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		flagDatabaseDSN = v
	}
	if v := os.Getenv("STORE_INTERVAL"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			flagStoreInterval = val
		}
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		flagFileStoragePath = v
	}
	if v := os.Getenv("RESTORE"); v != "" {
		if val, err := strconv.ParseBool(v); err == nil {
			flagRestore = val
		}
	}
	if v := os.Getenv("KEY"); v != "" {
		flagKey = v
	}
	if v := os.Getenv("CRYPTO_KEY"); v != "" {
		flagCryptoKey = v
	}
	if v := os.Getenv("CONFIG"); v != "" {
		flagConfigPath = v
	}
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		flagTrustedSubnet = v
	}
	return nil
}

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
		ServerAddress   *string `json:"server_address,omitempty"`
		DatabaseDSN     *string `json:"database_dsn,omitempty"`
		StoreInterval   *int    `json:"store_interval,omitempty"`
		FileStoragePath *string `json:"file_storage_path,omitempty"`
		Restore         *bool   `json:"restore,omitempty"`
		Key             *string `json:"key,omitempty"`
		CryptoKey       *string `json:"crypto_key,omitempty"`
		TrustedSubnet   *string `json:"trusted_subnet,omitempty"`
	}{}

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return err
	}

	if cfg.ServerAddress != nil {
		flagServerAddress = *cfg.ServerAddress
	}
	if cfg.DatabaseDSN != nil {
		flagDatabaseDSN = *cfg.DatabaseDSN
	}
	if cfg.StoreInterval != nil {
		flagStoreInterval = *cfg.StoreInterval
	}
	if cfg.FileStoragePath != nil {
		flagFileStoragePath = *cfg.FileStoragePath
	}
	if cfg.Restore != nil {
		flagRestore = *cfg.Restore
	}
	if cfg.Key != nil {
		flagKey = *cfg.Key
	}
	if cfg.CryptoKey != nil {
		flagCryptoKey = *cfg.CryptoKey
	}
	if cfg.TrustedSubnet != nil {
		flagTrustedSubnet = *cfg.TrustedSubnet
	}

	return nil
}
