package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/apps"
	"github.com/spf13/pflag"
)

// main is the program entry point.
//
// It prints build information, parses command-line flags, config file, and environment variables,
// and then starts the agent server using the parsed configuration options.
//
// The function panics if flag parsing, config loading, or server startup fails.
func main() {
	printBuildInfo()

	if err := parseFlags(); err != nil {
		panic(err)
	}

	if err := parseConfigFile(); err != nil {
		panic(err)
	}

	if err := parseEnvs(); err != nil {
		panic(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

var (
	flagServerAddress   string // address and port to run server
	flagDatabaseDSN     string // dsn for database connection
	flagStoreInterval   int    // interval (in seconds) to store data
	flagFileStoragePath string // path to store files
	flagRestore         bool   // whether to restore data from backup
	flagKey             string // key used for SHA256 hashing
	flagCryptoKey       string // path to private key file for encryption
	flagConfigPath      string // path to config file
	flagTrustedSubnet   string // trusted subnet in CIDR notation
	flagHashHeader      string // header for SHA256 hash
	flagLogLevel        string // log level for the application
	flagMigrationsDir   string // directory containing DB migration files
)

// parseFlags parses command-line flags and stores their values in package-level variables.
func parseFlags() error {
	pflag.StringVarP(&flagServerAddress, "address", "a", ":8080", "address and port to run server")
	pflag.StringVarP(&flagDatabaseDSN, "dsn", "d", "", "dsn for database connection")
	pflag.IntVarP(&flagStoreInterval, "interval", "i", 300, "interval (in seconds) to store data")
	pflag.StringVarP(&flagFileStoragePath, "file", "f", "", "path to store files")
	pflag.BoolVarP(&flagRestore, "restore", "r", false, "whether to restore data from backup")
	pflag.StringVarP(&flagKey, "key", "k", "", "key used for SHA256 hashing")
	pflag.StringVar(&flagCryptoKey, "crypto-key", "", "path to private key file for encryption")
	pflag.StringVarP(&flagConfigPath, "config", "c", "", "path to config file")
	pflag.StringVarP(&flagTrustedSubnet, "trusted-subnet", "t", "", "trusted subnet in CIDR notation")
	pflag.StringVar(&flagHashHeader, "hash-header", "H", "header for SHA256 hash")
	pflag.StringVarP(&flagLogLevel, "log-level", "l", "info", "log level for the application")
	pflag.StringVarP(&flagMigrationsDir, "migrations-dir", "m", "../../migrations", "directory containing DB migration files")

	pflag.Parse()

	return nil
}

// parseConfigFile loads configuration from a JSON file if the config path is set.
// It overrides flag variables with values from the config file.
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
		HashHeader      *string `json:"hash_header,omitempty"`
		LogLevel        *string `json:"log_level,omitempty"`
		MigrationsDir   *string `json:"migrations_dir,omitempty"`
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
	if cfg.HashHeader != nil {
		flagHashHeader = *cfg.HashHeader
	}
	if cfg.LogLevel != nil {
		flagLogLevel = *cfg.LogLevel
	}
	if cfg.MigrationsDir != nil {
		flagMigrationsDir = *cfg.MigrationsDir
	}

	return nil
}

// parseEnvs loads configuration from environment variables.
// It overrides flag variables with values from environment variables if set.
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
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		flagTrustedSubnet = v
	}
	if v := os.Getenv("HASH_HEADER"); v != "" {
		flagHashHeader = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		flagLogLevel = v
	}
	if v := os.Getenv("MIGRATIONS_DIR"); v != "" {
		flagMigrationsDir = v
	}

	return nil
}

// run initializes the server app with the parsed configuration and starts it.
func run() error {
	app, err := apps.NewServerApp(
		apps.WithServerAddress(flagServerAddress),
		apps.WithServerDatabaseDSN(flagDatabaseDSN),
		apps.WithServerStoreInterval(flagStoreInterval),
		apps.WithServerFileStoragePath(flagFileStoragePath),
		apps.WithServerRestore(flagRestore),
		apps.WithServerKey(flagKey),
		apps.WithServerCryptoKey(flagCryptoKey),
		apps.WithServerConfigPath(flagConfigPath),
		apps.WithServerTrustedSubnet(flagTrustedSubnet),
		apps.WithServerHashHeader(flagHashHeader),
		apps.WithServerLogLevel(flagLogLevel),
		apps.WithServerMigrationsDir(flagMigrationsDir),
	)

	if err != nil {
		return err
	}

	return app.Run(context.Background())
}

// printBuildInfo prints the build version, date, and commit hash to stdout.
// If any of these values are empty, it prints "N/A" instead.
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
