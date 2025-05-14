package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagServerAddress   string
	flagLogLevel        string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagDatabaseDSN     string
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, "a", ":8080", "Address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")
	flag.IntVar(&flagStoreInterval, "i", 300, "Interval (in seconds) to save server state to disk (default 300, 0 for synchronous)")
	flag.StringVar(&flagFileStoragePath, "f", "", "File path to save the server state (default empty, specify the file path using this flag)")
	flag.BoolVar(&flagRestore, "r", false, "Whether to restore server state from file (true/false)")
	flag.StringVar(&flagDatabaseDSN, "d", "", "PostgreSQL DSN (e.g., postgres://user:pass@localhost:5432/dbname)") // Новый флаг

	flag.Parse()

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		flagServerAddress = envServerAddress
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if interval, err := strconv.Atoi(envStoreInterval); err == nil {
			flagStoreInterval = interval
		}
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flagFileStoragePath = envFileStoragePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restore, err := strconv.ParseBool(envRestore); err == nil {
			flagRestore = restore
		}
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" && flagDatabaseDSN == "" {
		flagDatabaseDSN = envDatabaseDSN
	}
}
