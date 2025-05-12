package main

import (
	"flag"
	"os"
)

var (
	flagServerAddress string
	flagLogLevel      string
)

func parseFlags() {
	flag.StringVar(&flagServerAddress, "a", ":8080", "Address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		flagServerAddress = envServerAddress
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

}
